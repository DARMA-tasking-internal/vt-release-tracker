package main

import (
  "fmt"
  "os"
  "os/exec"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "strings"
  "sort"
  "regexp"
)

const url  = "https://api.github.com/repos/"
const org  = "DARMA-tasking"
const repo = "vt"
const base = url + org + "/" + repo
const git  = "git"

type IssueList struct {
  List        []*Issue       `json:"offer"`
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
  Id          int64          `json:"id"`
  Name        string         `json:"name"`
  Url         string         `json:"url"`
  Desc        string         `json:"description"`
}

type Milestone struct {
  Id          int64          `json:"id"`
  Number      int64          `json:"number"`
  Title       string         `json:"title"`
  State       string         `json:"state"`
  Url         string         `json:"url"`
  Desc        string         `json:"description"`
}

type Assignee struct {
  Id          int64          `json:"id"`
  Login       string         `json:"login"`
  Url         string         `json:"url"`
}

type PullRequstInfo struct {
  Url         string         `json:"url"`
}

type LabelMap map[string][]*Issue

type BranchMap map[int64]string

type BranchInfo struct {
  Merged   BranchMap
  Unmerged BranchMap
}

func main() {
  var args []string = os.Args

  if len(args) < 2 {
    fmt.Fprintf(os.Stderr, "usage: " + args[0] + " <branch> <release-tag>\n")
    os.Exit(1);
  }

  var branch, tag = args[1], args[2]
  fmt.Println("Analyzing branch:", branch)
  fmt.Println("Analyzing tag:", tag)

  var info = processRepo(tag)

  for issue, name := range info.Merged {
    fmt.Println("Merged:", issue, name)
  }
  for issue, name := range info.Unmerged {
    fmt.Println("Unmerged:", issue, name)
  }
}

func processRepo(ref string) *BranchInfo {
  const rp = "vt-base-repo"
  const uri = git + "@" + "github.com" + ":" + org + "/" + repo + ".git"

  if _, err := os.Stat(rp); os.IsNotExist(err) {
    _, err := exec.Command(git, "clone", uri, rp).Output()

    if err != nil {
      fmt.Fprintln(os.Stderr, "There was an error running git clone command: ", err)
      os.Exit(10)
    }
  }

  var rev = getRev(ref, rp)
  var info = new(BranchInfo)
  info.Merged   = branchMap(rev, rp, "--merged")
  info.Unmerged = branchMap(rev, rp, "--no-merged")
  return info
}

func getRev(ref string, rp string) string {
  var out, err = exec.Command(git, "-C", rp, "rev-parse", ref).Output()

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error running git rev-parse command", err)
    os.Exit(11)
  }

  return strings.Split(string(out), "\n")[0]
}

func branchMap(ref string, rp string, cmd string) BranchMap {
  var out, err = exec.Command(git, "-C", rp, "branch", "-r", cmd, ref).Output()

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error running git branch command", err)
    os.Exit(11)
  }

  var branches = string(out)
  var branch_list = strings.Split(branches, "\n")

  var branch_map = make(BranchMap)

  for _, branch := range branch_list {
    branch = strings.TrimSpace(branch)
    if branch != "" {
      name := strings.Split(branch, "origin/")[1]
      re := regexp.MustCompile(`(^[\d]+)-`)
      bname := []byte(name)
      if (re.Match(bname)) {
        issue := string(re.FindSubmatch(bname)[1])
        issue_num, _ := strconv.ParseInt(issue, 10, 64)
        branch_map[issue_num] = name
      }
    }
  }

  return branch_map
}

func processIssues() {
  var allIssues *IssueList = new(IssueList)
  npages := 7
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
}

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
  var row_len = 60
  var row_format = "| %-20v | %-6v | %-6v | %-6v | %-6v |\n";
  var row_div = strings.Repeat("-", row_len)
  fmt.Printf("%v\n", row_div)
  fmt.Printf(row_format, "Label", "Issues", "PRs", "Closed", "Total")
  fmt.Printf("%v\n", row_div)
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
    fmt.Printf(row_format, label, nissues, nprs, nclosed, len(issues))
  }
  fmt.Printf("%v\n", row_div)
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
