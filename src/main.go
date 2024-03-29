package main

import (
  "fmt"
  "os"
  "strconv"
  "strings"
)

const url  = "https://api.github.com/repos/"
const org  = "DARMA-tasking"
const repo = "vt"
const base = url + org + "/" + repo
const git  = "git"
const grep = "grep"

var analyzed_tag string = ""

func main() {
  var args []string = os.Args

  if len(args) < 2 {
    fmt.Fprintf(os.Stderr, "usage: " + args[0] + " <branch/tag> <label>...<label>\n")
    os.Exit(1);
  }

  var tag = args[1]
  var labels []string
  for i := 2; i < len(args); i++ {
    labels = append(labels, args[i])
  }
  analyzed_tag = tag

  fmt.Println("Analyzing tag/branch in repository:", tag)
  var info = processRepo(tag)

  fmt.Print("Merged branches: " + tag + ": ")
  for issue, _ := range info.Merged {
    fmt.Print(issue, " ")
  }
  fmt.Print("\nUnmerged branches: " + tag + ": ")
  for issue, _ := range info.Unmerged {
    fmt.Print(issue, " ")
  }
  fmt.Print("\n")

  fmt.Println("Analyzing labels on Github:", labels)
  var label_map, all = processIssues()
  var lookupOnLabel = make(IssueOnLabelMap)
  for _, l := range labels {
    for _, issue := range label_map[l] {
      addOnLabel(&lookupOnLabel, issue, l)
    }
  }

  var state = buildState(lookupOnLabel, info, all)

  var mcorrect    = "\033[32m"   + "Merged Correctly    " + "\033[00m"
  var mincorrect  = "\033[91m"   + "Merged Incorrectly  " + "\033[00m"
  var umincorrect = "\033[31;1m" + "Unmerged Incorrectly" + "\033[00m"
  var umcorrect   = "\033[34m"   + "Unmerged Correctly  " + "\033[00m"
  var nobranch    = "\033[33m"   + "Unmerged No Branch! " + "\033[00m"

  printTable(MergedOnLabel,    mcorrect,    lookupOnLabel, state, true)
  printTable(MergedOffLabel,   mincorrect,  lookupOnLabel, state, false)
  printTable(UnmergedOnLabel,  umincorrect, lookupOnLabel, state, false)
  printTable(UnmergedOffLabel, umcorrect,   lookupOnLabel, state, false)
  printTable(UnmergedNoBranch, nobranch,    lookupOnLabel, state, false)
}

func buildState(lookupOnLabel IssueOnLabelMap, info *BranchInfo, all *IssueList) MergeStateMap {
  var state = make(MergeStateMap)
  state[MergedOnLabel]    = make(MergeIssueMap)
  state[MergedOffLabel]   = make(MergeIssueMap)
  state[UnmergedOnLabel]  = make(MergeIssueMap)
  state[UnmergedOffLabel] = make(MergeIssueMap)
  state[UnmergedNoBranch] = make(MergeIssueMap)

  for issue, branch := range info.Merged {
    if lookupOnLabel[issue] == nil {
      var on = MergedOffLabel
      state[on][issue] = createMergeState(issue, branch, on, all)
    } else {
      var on = MergedOnLabel
      state[on][issue] = createMergeState(issue, branch, on, all)
    }
  }

  for issue, branch := range info.Unmerged {
    if lookupOnLabel[issue] == nil {
      var on = UnmergedOffLabel
      state[on][issue] = createMergeState(issue, branch, on, all)
    } else {
      var on = UnmergedOnLabel
      state[on][issue] = createMergeState(issue, branch, on, all)
    }
  }

  for _, elm := range lookupOnLabel {
    var i *Issue = nil
    if elm.Issue.IsPR {
      i = elm.Issue.PRIssue;
    } else {
      i = elm.Issue;
    }
    if i != nil {
      if info.Merged[i.Number] == "" && info.Unmerged[i.Number] == "" {
        var on = UnmergedNoBranch
        state[on][i.Number] = createMergeState(i.Number, "", on, all)
      }
    }
  }

  return state
}

func printTable(key int, status string, lookup IssueOnLabelMap, state MergeStateMap, header bool) {
  var row_len = 161
  var row_format = "| %-6v | %-6v | %-20v | %-15v | %-45v | %-50v |\n";
  var row_div = strings.Repeat("-", row_len)
  if header {
    fmt.Printf("%v\n", row_div)
    var title = "RESULTS FROM ANALYSIS OF BRANCH \"" + analyzed_tag + "\""
    var spacer = strings.Repeat(" ", ((row_len/2) - (len(title)/2)))
    fmt.Printf(" %v%v%v\n", spacer, title, spacer)
    fmt.Printf("%v\n", row_div)
    fmt.Printf(row_format, "Issue", "PR", "State", "Issue State", "Branch", "Matching labels")
    fmt.Printf("%v\n", row_div)
  }
  if len(state[key]) == 0 {
    fmt.Printf(" %v\n", "No issues " + status)
  }
  for _, st := range state[key] {
    var num = st.Issue.Number
    var branch = st.BranchName
    var istate = st.Issue.State
    var pr = ""
    var labels []string
    if lookup[num] != nil {
      labels = lookup[num].Labels
    }
    var label_str = ""
    for i, l := range labels {
      label_str += l
      if i < len(labels)-1 {
        label_str += ", "
      }
    }
    if st.Issue.PRIssue != nil {
      pr = strconv.FormatInt(st.Issue.PRIssue.Number, 10)
    }
    fmt.Printf(row_format, num, pr, status, istate, branch, label_str)
  }
  fmt.Printf("%v\n", row_div)
}

func findIssue(issue int64, all *IssueList) *Issue {
  for _, i := range all.List {
    if i.Number == issue {
      return i
    }
  }
  return nil
}

func createMergeState(issue int64, branch string, state int, all *IssueList) *MergeState {
  var ms = new(MergeState)
  ms.Issue = findIssue(issue, all)
  if ms.Issue != nil {
    ms.PR = ms.Issue.PRIssue
  }
  ms.BranchName = branch
  ms.State = state
  return ms
}

func addOnLabel(on *IssueOnLabelMap, issue *Issue, label string) {
  if (*on)[issue.Number] == nil {
    var elm = new(IssueOnLabels)
    elm.Issue = issue
    elm.Labels = append(elm.Labels, label)
    (*on)[issue.Number] = elm
  } else {
    (*on)[issue.Number].Labels = append((*on)[issue.Number].Labels, label)
  }
}
