package main

import (
  "fmt"
  "os"
  "io"
  "bytes"
  "os/exec"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "strings"
  "sort"
  "regexp"
  // "log"
)

const url  = "https://api.github.com/repos/"
const org  = "DARMA-tasking"
const repo = "vt"
const base = url + org + "/" + repo
const git  = "git"
const grep = "grep"

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
  PRIssue     *Issue
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
    fmt.Fprintf(os.Stderr, "usage: " + args[0] + " <label> <branch/tag>\n")
    os.Exit(1);
  }

  var label, tag = args[1], args[2]

  fmt.Println("Analyzing tag/branch in repository:", tag)
  var info = processRepo(tag)

  fmt.Print("Merged into: " + tag + ": ")
  for issue, _ := range info.Merged {
    fmt.Print(issue, " ")
  }
  fmt.Print("\nNot merged into: " + tag + ": ")
  for issue, _ := range info.Unmerged {
    fmt.Print(issue, " ")
  }
  fmt.Print("\n")

  fmt.Println("Analyzing label on Github:", label)
  var labels = processIssues()
  var lookup = make(map[int64]*Issue)
  for _, issue := range labels[label] {
    lookup[issue.Number] = issue
  }
  for _, issue := range labels[label] {
    if !issue.IsPR {
      var in_merged, in_unmerged bool
      in_merged = info.Merged[issue.Number] != ""
      in_unmerged = info.Unmerged[issue.Number] != ""
      fmt.Println("issue:", issue.Number, " merged=", in_merged, " unmerged=", in_unmerged)
    } else {
      if issue.PRIssue != nil {
        if lookup[issue.PRIssue.Number] == nil {
          fmt.Print("PR #", issue.Number, ": has label=", label)
          fmt.Print(", but issue #", issue.PRIssue.Number, " does not\n")
          //addLabelToIssue(issue.PRIssue.Number, label)
        }
      }
    }
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

  var info2 = new(BranchInfo)
  info2.Merged = info.Merged
  info2.Unmerged = make(BranchMap)

  for issue_num, branch_name := range info.Unmerged {
    //fmt.Println("issue_num=", issue_num, "branch_name=", branch_name)
    found, _ := grepLogCheckMerge(ref, rp, strconv.FormatInt(issue_num, 10))
    if found {
      //fmt.Println("branch=", branch_name, "found=", found)
      info2.Merged[issue_num] = branch_name
    } else {
      info2.Unmerged[issue_num] = branch_name
    }
  }

  return info2
}

func getRev(ref string, rp string) string {
  var out, err = exec.Command(git, "-C", rp, "rev-parse", ref).Output()

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error running git rev-parse command", err)
    os.Exit(11)
  }

  return strings.Split(string(out), "\n")[0]
}

func grepLogCheckMerge(ref string, rp string, issue string) (bool, []string) {
  var c1 = exec.Command(git, "-C", rp, "log", ref)
  var c2 = exec.Command(grep, "#" + issue)

  r, w := io.Pipe()
  c1.Stdout = w
  c2.Stdin = r

  var b2 bytes.Buffer
  c2.Stdout = &b2

  c1.Start()
  c2.Start()
  c1.Wait()
  w.Close()
  c2.Wait()

  var commits = strings.Split(b2.String(), "\n")
  var cleaned []string
  for _, commit := range commits {
    commit = strings.TrimSpace(commit)
    if (commit != "") {
      //fmt.Println("grep: out=", commit)
      cleaned = append(cleaned, commit)
    }
  }

  var found bool = len(cleaned) > 0
  return found, cleaned
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

func processIssues() LabelMap {
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

  var lookup = make(map[int64]*Issue)
  apply(allIssues, func(i *Issue) { lookup[i.Number] = i; })
  apply(allIssues, func(i *Issue) { i.IsPR = i.PullRequest.Url != ""; })
  apply(allIssues, func(i *Issue) {
    if i.IsPR {
      var arr = strings.Split(i.Title, " ")
      if len(arr) > 0 {
        x, err := strconv.ParseInt(arr[0], 10, 64)
        if err == nil {
          if lookup[x] != nil {
            //fmt.Print("PR=", i.Number, ": found issue: ", x, "\n")
            i.PRIssue = lookup[x]
          }
        } else {
          //fmt.Print("PR=", i.Number, ": does not follow format: title=", i.Title, "\n")
        }
      }
    }
  })

  labels := makeLabelMap(allIssues)

  printBreakdown(labels)

  return labels
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

/*
 * This requires more oauth2 than I want to implement right now, so these
 * functions do not work for modifying the without generating a personal
 * access token
 */

func buildPost(element string, query map[string]string) string {
  var target string = base + "/" + element;
  var paged string = target
  query["client_id"]     = "68bbbfd492795c4835bb"
  query["client_secret"] = "7ecf3be133774fe30076747cdc09fd65e23d2cf8"
  var i = 0
  for key, val := range query {
    if i == 0 {
      paged += "?" + key + "=" + val
    } else {
      paged += "&" + key + "=" + val
    }
    i++
  }
  return paged
}

func addLabelToIssue(issue int64, label string) bool {
  var issue_str = strconv.FormatInt(issue, 10)
  var query = make(map[string]string)
  var target = buildPost("issues/" + issue_str + "/labels", query)

  fmt.Println("addLabelToIssue: id=", issue, ", label=", label)
  fmt.Println("addLabelToIssue: target=", target)

  var json = []byte(`{"labels": ["` + label + `"]}`)

  response, err := http.Post(target, "application/json", bytes.NewBuffer(json))
  if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to post target:" + target + "\n")
    return false
  }

  body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("post:\n", string(body))

  defer response.Body.Close()
  return true
}
