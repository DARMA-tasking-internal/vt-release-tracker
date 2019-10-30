package main

import (
  "fmt"
  "os"
  "os/exec"
  "strconv"
  "strings"
  "regexp"
)

func getBranches() []*BranchName {
  const rp = "vt-base-repo"
  const uri = git + "@" + "github.com" + ":" + org + "/" + repo + ".git"

  if _, err := os.Stat(rp); os.IsNotExist(err) {
    _, err := exec.Command(git, "clone", uri, rp).Output()

    if err != nil {
      fmt.Fprintln(os.Stderr, "There was an error running git clone command: ", err)
      os.Exit(10)
    }
  } else {
    _, err := exec.Command(git, "-C", rp, "pull", "--all").Output()

    if err != nil {
      fmt.Fprintln(os.Stderr, "There was an error running git pull command: ", err)
      os.Exit(10)
    }
  }

  var list []*BranchName

  {
    var out, err = exec.Command(git, "-C", rp, "tag", "-l").Output()

    if err != nil {
      fmt.Fprintln(os.Stderr, "Error running git tag -l command", err)
      os.Exit(11)
    }

    for _, name := range strings.Split(string(out), "\n") {
      name = strings.TrimSpace(name)
      if (name != "") {
        var bn = new(BranchName)
        bn.Branch = name
        list = append(list, bn)
      }
    }
  }

  {
    var out, err = exec.Command(git, "-C", rp, "branch", "-r").Output()

    if err != nil {
      fmt.Fprintln(os.Stderr, "Error running git branch command", err)
      os.Exit(11)
    }

    for _, name := range strings.Split(string(out), "\n") {
      name = strings.TrimSpace(name)
      if (name != "") {
        var bn = new(BranchName)
        bn.Branch = name
        list = append(list, bn)
      }
    }
  }

  return list
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
  } else {
    _, err := exec.Command(git, "-C", rp, "pull", "--all").Output()

    if err != nil {
      fmt.Fprintln(os.Stderr, "There was an error running git pull command: ", err)
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

  for issue_num, how := range info.Unmerged {
    var branch_name = how.Branch
    //fmt.Println("issue_num=", issue_num, "branch_name=", branch_name)
    var issue_str = strconv.FormatInt(issue_num, 10)
    found, _ := grepLogCheckMerge(ref, rp, issue_str, "#" + issue_str)
    if found {
      info2.Merged[issue_num] = branchFound(branch_name, IssueGrep)
    } else {
      found2, _ := grepLogMessage("origin/" + branch_name, ref, rp, issue_str)
      //fmt.Println("branch=", branch_name, "found2=", found2)
      if found2 {
        info2.Merged[issue_num] = branchFound(branch_name, CommitGrepMsg)
      } else {
        info2.Unmerged[issue_num] = branchFound(branch_name, Merged)
      }
    }
  }

  return info2
}

func branchFound(branch string, how int) *BranchFound {
  var bf = new(BranchFound)
  bf.Branch = branch
  bf.How = how
  return bf
}

func getRev(ref string, rp string) string {
  var out, err = exec.Command(git, "-C", rp, "rev-parse", ref).Output()

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error running git rev-parse command", err)
    os.Exit(11)
  }

  return strings.Split(string(out), "\n")[0]
}

func grepLogCheckMerge(ref string, rp string, issue string, pattern string) (bool, []string) {
  //var str = fmt.Sprintf("--grep=\"%s\"", pattern)
  var cmd = []string{"-C", rp, "log", ref, "--oneline", "--grep", pattern}
  var out, err = exec.Command(git, cmd...).Output()

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error running git log command", err)
    os.Exit(11)
  }

  var out_str = string(out)
  var commits = strings.Split(out_str, "\n")
  var cleaned []string
  for _, commit := range commits {
    commit = strings.TrimSpace(commit)
    if (commit != "") {
      cleaned = append(cleaned, commit)
    }
  }

  var found bool = len(cleaned) > 0
  return found, cleaned
}

func grepLogMessage(branch string, ref string, rp string, issue string) (bool, []string) {
  var cmd = []string{"-C", rp, "log", "--no-decorate", "--sparse", "--format=%B", "-n", "1", branch}
  var out, err = exec.Command(git, cmd...).Output()

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error running git log command", err)
    os.Exit(11)
  }

  var msg_multi = string(out)
  var msg_array = strings.Split(msg_multi, "\n")
  var msg = msg_array[0]

  return grepLogCheckMerge(ref, rp, issue, msg)
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
        branch_map[issue_num] = branchFound(name, Merged)
      } else {
        re := regexp.MustCompile(`^feature-([\d]+)-`)
        if (re.Match(bname)) {
          issue := string(re.FindSubmatch(bname)[1])
          issue_num, _ := strconv.ParseInt(issue, 10, 64)
          branch_map[issue_num] = branchFound(name, Merged)
        }
      }
    }
  }

  return branch_map
}

