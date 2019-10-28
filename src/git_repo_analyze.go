package main

import (
  "fmt"
  "os"
  "io"
  "bytes"
  "os/exec"
  "strconv"
  "strings"
  "regexp"
)

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

  fetchBranch(ref, rp)
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

func fetchBranch(ref string, rp string) {
  var ref2 = ref + ":" + ref
  cmd := exec.Command(git, "-C", rp, "fetch", "origin", ref2)
  var out bytes.Buffer
  var stderr bytes.Buffer
  cmd.Stdout = &out
  cmd.Stderr = &stderr
  err := cmd.Run()
  if err != nil {
    fmt.Fprintln(os.Stderr, fmt.Sprint(err) + ": " + stderr.String() + "when running git fetch command")
    os.Exit(11);
  }
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

