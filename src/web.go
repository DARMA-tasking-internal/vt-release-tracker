package main

import (
  "fmt"
  "strings"
  "log"
  "net/http"
  "html/template"
)

func startWebServer() {
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
  http.HandleFunc("/vt", selectIssuesBranch)
  http.HandleFunc("/analyze", analyze)
  http.HandleFunc("/analyzeUrl", analyzeUrl)

  log.Fatal(http.ListenAndServe(":8080", nil))
}

func selectIssuesBranch(w http.ResponseWriter, r *http.Request) {
  var label_map, _, _ = processIssues()
  var table = makeTable(label_map)
  var branches = getBranches()
  table.BranchList = branches

  t, _ := template.ParseFiles("issues.html")
  t.Execute(w, table)
}

func analyzeUrl(w http.ResponseWriter, r *http.Request) {
  var branch string = ""
  var labels []string

  tag,       has_tag   := r.URL.Query()["tag"]
  label_map, has_label := r.URL.Query()["label"]

  //var uri = r.URL.Path[1:]
  fmt.Println("tag=", tag)
  fmt.Println("label_map=", label_map)

  if !has_tag || len(tag[0]) < 1 {
    fmt.Println("missing tag")
    return
  }

  if !has_label || len(label_map[0]) < 1 {
    fmt.Println("missing labels")
    return
  }

  branch = tag[0]
  labels = label_map

  if len(labels) > 0 && branch != "" {
    doAnalyze(w, branch, labels);
  }
}

func analyze(w http.ResponseWriter, r *http.Request) {
  var tag string = ""
  var labels []string

  r.ParseForm()
  fmt.Println(r.Form)
  fmt.Println("path", r.URL.Path)
  fmt.Println("scheme", r.URL.Scheme)
  fmt.Println(r.Form["url_long"])
  for k, v := range r.Form {
    if k == "branch" {
      tag = strings.Join(v, "")
    } else if k == "labels" {
      var label_concat = strings.Join(v, "")
      var label_list = strings.Split(label_concat, ";")
      for _, l := range label_list {
        l = strings.TrimSpace(l)
        if (l != "") {
          labels = append(labels, l)
        }
      }
    } else {
      fmt.Println("key:", k)
      fmt.Println("val:", strings.Join(v, ""))
    }
  }

  doAnalyze(w, tag, labels)
}

func doAnalyze(w http.ResponseWriter, tag string, labels []string) {
  fmt.Println("tag:", tag)
  fmt.Println("labels", labels)

  var info = processRepo(tag)
  var rev = getRev(tag, "vt-base-repo")
  var label_map, label_data, all = processIssues()
  var lookupOnLabel = make(IssueOnLabelMap)
  for _, l := range labels {
    for _, issue := range label_map[l] {
      addOnLabel(&lookupOnLabel, issue, l)
    }
  }

  var state = buildState(lookupOnLabel, info, all)

  var mcorrect    = "merged correctly"
  var mincorrect  = "merged incorrectly"
  var umincorrect = "unmerged incorrectly"
  var umcorrect   = "unmerged correctly"
  var nobranch    = "nobranch"

  var t1 = makeMergeStatus(MergedOnLabel,    mcorrect,    lookupOnLabel, state, label_data)
  var t2 = makeMergeStatus(MergedOffLabel,   mincorrect,  lookupOnLabel, state, label_data)
  var t3 = makeMergeStatus(UnmergedOnLabel,  umincorrect, lookupOnLabel, state, label_data)
  var t4 = makeMergeStatus(UnmergedOffLabel, umcorrect,   lookupOnLabel, state, label_data)
  var t5 = makeMergeStatus(UnmergedNoBranch, nobranch,    lookupOnLabel, state, label_data)

  var table = new(MergeStatusTable)
  table = addRowTable(table, t2, mincorrect)
  table = addRowTable(table, t3, umincorrect)
  table = addRowTable(table, t5, nobranch)
  table = addRowTable(table, t1, mcorrect)
  table = addRowTable(table, t4, umcorrect)

  var progress_total int     = 0
  var merged_on_num          = len(state[MergedOnLabel])
  var merged_off_num         = len(state[MergedOffLabel])
  var unmerged_on_num        = len(state[UnmergedOnLabel])
  var unmerged_off_num       = len(state[UnmergedOffLabel])
  var unmerged_no_branch_num = len(state[UnmergedNoBranch])

  progress_total += merged_on_num
  progress_total += merged_off_num
  progress_total += unmerged_no_branch_num
  progress_total += unmerged_on_num
  progress_total += unmerged_off_num

  var merged_on_percent          = float32(merged_on_num)          / float32(progress_total) * 100.0
  var merged_off_percent         = float32(merged_off_num)         / float32(progress_total) * 100.0
  var unmerged_on_percent        = float32(unmerged_on_num)        / float32(progress_total) * 100.0
  var unmerged_off_percent       = float32(unmerged_off_num)       / float32(progress_total) * 100.0
  var unmerged_no_branch_percent = float32(unmerged_no_branch_num) / float32(progress_total) * 100.0

  table.MergedOnProgress         = merged_on_percent
  table.MergedOffProgress        = merged_off_percent
  table.UnmergedNoBranchProgress = unmerged_no_branch_percent
  table.UnmergedOnProgress       = unmerged_on_percent
  table.UnmergedOffProgress      = unmerged_off_percent

  fmt.Println("MergedOnLabel=", merged_on_num);
  fmt.Println("MergedOffLabel=", merged_off_num);
  fmt.Println("UnmergedOnLabel=", unmerged_on_num);
  fmt.Println("UnmergedNoBranch=", unmerged_no_branch_num);
  fmt.Println("total=", progress_total);

  table.Branch = tag
  table.Rev = rev
  for _, l := range labels {
    var label = new(LabelName)
    label.Label = l
    label.Color = label_data[l].Color
    label.Url = label_data[l].Url
    table.LabelList = append(table.LabelList, label)
  }
  table.Url = "analyzeUrl?tag=" + tag + "&label=" + strings.Join(labels, "&label=")
  t, _ := template.ParseFiles("merged.html")
  t.Execute(w, table)
}

func addRowTable(table *MergeStatusTable, t1 *MergeStatusTable, status string) *MergeStatusTable {
  var found = len(t1.List) != 0

  for _, e := range t1.List {
    table.List = append(table.List, e)
  }

  var entry = new(MergeStatus)
  entry.Spacer = true
  entry.SpacerStatus = found
  entry.Status = status
  table.List = append(table.List, entry)

  return table
}
