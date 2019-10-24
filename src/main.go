package main

import (
  "fmt"
  "os"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
  //"strings"
)

type IssueList struct {
  List []Issue `json:"offer"`
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

  apply(allIssues, func(i *Issue) { i.IsPR = i.PullRequest.Url != ""; })

  var labels = make(LabelMap)
  apply(allIssues, func(i *Issue) {
    for _, l := range i.Labels {
      labels[l.Name] = append(labels[l.Name], i)
    }
  })

  for label, issues := range labels {
    fmt.Println("label=" + label + ": issues=" + strconv.Itoa(len(issues)))
  }

  //processIssues(allIssues)
}

const url  = "https://api.github.com/repos/"
const base = url + "DARMA-tasking/vt"

func buildGet(element string, page int, query map[string]string) string {
  var target string = base + "/" + element;
  var paged string = target + "?page=" + strconv.Itoa(page)
  //query["per_page"] = "100"
  for key, val := range query {
    paged += "&" + key + "=" + val
  }
  return paged
}

func apply (issues *IssueList, fn func(*Issue)) {
  for i, _ := range issues.List {
    fn(&issues.List[i])
  }
}

// func processIssues(issues *IssueList) {
//   var label_map = make(map[string][]Issue)

//   for _, issue := range issues.List {
//     var issueNumStr = strconv.FormatInt(issue.Number, 10)
//     fmt.Println("id=" + issueNumStr + ", " + "labels " + strconv.Itoa(len(issue.Labels)));


//     //issue.IsPR = issue.PullRequest.Url != ""
//     fmt.Println("is_pr=", issue.IsPR)

//     // var sliced = strings.Split(issue.PullRequest.Url, "/")
//     // var pr_str = sliced[len(sliced)-1]
//     // var pr_num, _ = strconv.Atoi(pr_str)
//     // var is_pr = int64(pr_num) == issue.Number

//     // fmt.Print("pr_num=", pr_str, ", is_pr=", is_pr, " url=", issue.PullRequest.Url, "\n")

//     for _, label := range issue.Labels {
//       var name string = label.Name
//       label_map[name] = append(label_map[name], issue)
//     }
//   }

//   for name, list := range label_map {
//     fmt.Println("label=" + name + ": issues=" + strconv.Itoa(len(list)))
//   }
// }

func getIssues(state string, page int, out chan<- *IssueList) {
  var query = make(map[string]string)
  query["state"] = state

  var target = buildGet("issues", page, query)
  fmt.Println("string is: " + target)

  response, err := http.Get(target)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to fetch target:" + target + "\n")
    os.Exit(3)
  }

  //fmt.Fprintf(os.Stderr, "response=%s\n\n", response)

  defer response.Body.Close()

  var issues *IssueList = new(IssueList)
  raw_data, _ := ioutil.ReadAll(response.Body)
  err = json.Unmarshal(raw_data, &issues.List)

  if err != nil {
    fmt.Println(err)
    fmt.Fprintf(os.Stderr, "failure parsing json response\n")
    os.Exit(2);
  }

  for _, issue := range issues.List {
    fmt.Println("id=" + strconv.FormatInt(issue.Id, 10) + ", number=" + strconv.FormatInt(issue.Number, 10))
  }

  var num_fetched string = strconv.Itoa(len(issues.List))
  fmt.Println("Fetched " + num_fetched + " issues")

  out <- issues
}