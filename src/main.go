package main

import (
  "fmt"
  "os"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "strings"
)

type IssueList struct {
  List []*Issue `json:"offer"`
}

type Issue struct {
  Id          int64          `json:"id"`
  Number      int64          `json:"number"`
  Title       string         `json:"title"`
  State       string         `json:"state"`
  Url         string         `json:"url"`
  Body        string         `json:"body"`
  Labels      []Label        `json:"labels"`
  Milestone   Milestone      `json:"milestone"`
  Assignee    []Assignee     `json:"assignees"`
  PullRequest PullRequstInfo `json:"pull_request"`
  IsPR        bool
}

type Label struct {
  Id      int64  `json:"id"`
  Name    string `json:"name"`
  Url     string `json:"url"`
  Desc    string `json:"description"`
}

type Milestone struct {
  Id      int64   `json:"id"`
  Number  int64   `json:"number"`
  Title   string  `json:"title"`
  State   string  `json:"state"`
  Url     string  `json:"url"`
  Desc    string ` json:"description"`
}

type Assignee struct {
  Id      int64   `json:"id"`
  Login   string  `json:"login"`
  Url     string  `json:"url"`
}

type PullRequstInfo struct {
  Url     string  `json:"url"`
}

type LabelMap map[string][]*Issue

func main() {
  var args []string = os.Args

  if len(args) < 2 {
    fmt.Fprintf(os.Stderr, "usage: " + args[0] + " <branch> <release-tag>\n")
    os.Exit(1);
  }

  var allIssues *IssueList = new(IssueList)
  npages := 1
  issueChannel := make(chan *IssueList, npages)

  for i := 0; i < npages; i++ {
    go getIssues("all", i, issueChannel)
  }

  for i := 0; i < npages; i++ {
    var issuePage *IssueList = <-issueChannel
    allIssues.List = append(allIssues.List, issuePage.List...)
  }

  fmt.Println("Fetched", len(allIssues.List), "issues")

  apply(allIssues, func(i *Issue) { i.IsPR = i.PullRequest.Url != ""; })

  labels := makeLabelMap(allIssues)
  printBreakdown(labels)

  //processIssues(allIssues)
}

const url  = "https://api.github.com/repos/"
const base = url + "DARMA-tasking/vt"

func makeLabelMap(issues *IssueList) LabelMap {
  var labels = make(LabelMap)
  apply(issues, func(i *Issue) {
    for _, l := range i.Labels {
      labels[l.Name] = append(labels[l.Name], i)
    }
  })
  return labels
}

func printBreakdown(labels LabelMap) {
  var row_format = "| %-20v | %-15v | %-15v | %-15v |\n";
  fmt.Printf("%v\n", strings.Repeat("-", 78))
  fmt.Printf(row_format, "Label", "Issues", "PRs", "Total")
  fmt.Printf("%v\n", strings.Repeat("-", 78))
  for label, issues := range labels {
    var nprs, nissues int
    var list = new(IssueList)
    list.List = issues
    apply(list, func(i *Issue) { if i.IsPR { nprs++ } else { nissues++ }; })
    fmt.Printf(row_format, label, nissues, nprs, len(issues))
  }
  fmt.Printf("%v\n", strings.Repeat("-", 78))
}

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

func apply(issues *IssueList, fn func(*Issue)) {
  for i, _ := range issues.List {
    fn(issues.List[i])
  }
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
