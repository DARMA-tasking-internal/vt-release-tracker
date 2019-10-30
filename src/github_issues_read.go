package main

import (
  "fmt"
  "os"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "strings"
  "sort"
  "regexp"
)

func buildGet(element string, page int, query map[string]string) string {
  var target string = base + "/" + element;
  var paged string = target + "?page=" + strconv.Itoa(page)
  query["per_page"]      = "100"
  query["client_id"]     = "68bbbfd492795c4835bb"
  query["client_secret"] = "7ecf3be133774fe30076747cdc09fd65e23d2cf8"
  for key, val := range query {
    paged += "&" + key + "=" + val
  }
  return paged
}

func getIssues(state string, page int, out chan<- *IssueList) {
  var query = make(map[string]string)
  query["state"] = state
  var target = buildGet("issues", page, query)

  response, err := http.Get(target)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to fetch target:" + target + "\n")
    os.Exit(3)
  }

  defer response.Body.Close()

  var issues *IssueList = new(IssueList)
  raw_data, _ := ioutil.ReadAll(response.Body)
  err = json.Unmarshal(raw_data, &issues.List)

  if err != nil {
    fmt.Println(err)
    fmt.Fprintf(os.Stderr, "failure parsing json response\n")
    os.Exit(2);
  }

  out <- issues
}

func apply(issues *IssueList, fn func(*Issue)) {
  for i, _ := range issues.List {
    fn(issues.List[i])
  }
}

func processIssues() (LabelMap, LabelData, *IssueList) {
  var allIssues *IssueList = new(IssueList)
  npages := 7
  issueChannel := make(chan *IssueList, npages)

  for i := 1; i < npages; i++ {
    go getIssues("all", i, issueChannel)
  }

  for i := 1; i < npages; i++ {
    var issuePage *IssueList = <-issueChannel
    allIssues.List = append(allIssues.List, issuePage.List...)
  }

  fmt.Println("Fetched", len(allIssues.List), "issues")

  var pr_to_issue = make(map[int64]*Issue)
  var issue_to_pr = make(map[int64]*Issue)
  var lookup = make(map[int64]*Issue)
  apply(allIssues, func(i *Issue) { lookup[i.Number] = i; })
  apply(allIssues, func(i *Issue) { i.IsPR = i.PullRequest.Url != ""; })
  apply(allIssues, func(i *Issue) {
    if i.IsPR {

      var match bool = false
      var found string = ""
      re := regexp.MustCompile(`^[#]?([\d]+)`)
      title_byte := []byte(i.Title)
      if (re.Match(title_byte)) {
        found = string(re.FindSubmatch(title_byte)[1])
        match = true
        fmt.Println("title=", i.Title, "found=", found)
      } else {
        re := regexp.MustCompile(`^Feature [#]?([\d]+)`)
        if (re.Match(title_byte)) {
          found = string(re.FindSubmatch(title_byte)[1])
          match = true
          fmt.Println("2 title=", i.Title, "found=", found)
        }
      }

      if match {
        x, err := strconv.ParseInt(found, 10, 64)
        if err == nil {
          if lookup[x] != nil {
            //fmt.Print("PR=", i.Number, ": found issue: ", x, "\n")
            i.PRIssue = lookup[x]
            pr_to_issue[i.Number] = i.PRIssue
          }
        } else {
          //fmt.Print("PR=", i.Number, ": does not follow format: title=", i.Title, "\n")
        }
      }
    }
  })

  for pr, issue := range pr_to_issue {
    issue_to_pr[issue.Number] = findIssue(pr, allIssues)
  }

  apply(allIssues, func(i *Issue) {
    if !i.IsPR {
      i.PRIssue = issue_to_pr[i.Number]
    }
  })

  labels, label_data := makeLabelMap(allIssues)

  printBreakdown(labels)

  return labels, label_data, allIssues
}

func makeLabelMap(issues *IssueList) (LabelMap, LabelData) {
  var labels = make(LabelMap)
  var label_data = make(LabelData)
  apply(issues, func(i *Issue) {
    for _, l := range i.Labels {
      labels[l.Name] = append(labels[l.Name], i)
      label_data[l.Name] = l
    }
  })
  return labels, label_data
}

func makeTable(labels LabelMap) *IssueTable {
  var table = new(IssueTable)

  var keys = make([]string, 0, len(labels))
  for label, _ := range labels {
    keys = append(keys, label)
  }
  sort.Strings(keys)

  for _, label := range keys {
    var issues = labels[label]
    var nprs, nissues, nclosed int
    var list = new(IssueList)
    list.List = issues
    apply(list, func(i *Issue) {
      if i.IsPR { nprs++ } else { nissues++ };
      if i.State == "closed" { nclosed++ }
    })
    var entry = &IssueTableEntry{
      Label: label,
      Issues: nissues,
      PRs: nprs,
      Closed: nclosed,
      Total: len(issues) }
    table.List = append(table.List, entry)
  }
  return table
}

func printBreakdown(labels LabelMap) {
  var table = makeTable(labels)
  var row_len = 60
  var row_format = "| %-20v | %-6v | %-6v | %-6v | %-6v |\n";
  var row_div = strings.Repeat("-", row_len)
  fmt.Printf("%v\n", row_div)
  fmt.Printf(row_format, "Label", "Issues", "PRs", "Closed", "Total")
  fmt.Printf("%v\n", row_div)
  for _, e := range table.List {
    fmt.Printf(row_format, e.Label, e.Issues, e.PRs, e.Closed, e.Total)
  }
  fmt.Printf("%v\n", row_div)
}

