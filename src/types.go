package main

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
  Color       string         `json:"color"`
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

type LabelMap    map[string][]*Issue
type LabelData   map[string]Label

const (
  Merged        = iota
  IssueGrep     = iota
  CommitGrepMsg = iota
)

type BranchFound struct {
  Branch   string
  How      int
}

type BranchMap map[int64]*BranchFound

type BranchInfo struct {
  Merged   BranchMap
  Unmerged BranchMap
}

const (
  MergedOnLabel    = iota
  MergedOffLabel   = iota
  UnmergedOnLabel  = iota
  UnmergedOffLabel = iota
  UnmergedNoBranch = iota
)

type MergeState struct {
  Issue       *Issue
  PR          *Issue
  BranchName  string
  State       int
  How         int
}

type IssueOnLabels struct {
  Issue       *Issue
  Labels      []string
  LabelData   []Label
}

type IssueOnLabelMap map[int64]*IssueOnLabels
type MergeIssueMap   map[int64]*MergeState
type MergeStateMap   map[int]MergeIssueMap

type IssueTable struct {
  List       []*IssueTableEntry
  BranchList []*BranchName
}

type BranchName struct {
  Branch     string
}

type IssueTableEntry struct {
  Label       string
  Issues      int
  PRs         int
  Closed      int
  Total       int
}

type MergeStatusTable struct {
  List      []*MergeStatus
  Branch    string
  LabelList string
  Url       string
}

type MergeStatus struct {
  Issue        string
  PR           string
  Status       string
  IssueStatus  string
  Branch       string
  Labels       []*LabelName
  Caveat       string
  Spacer       bool
  SpacerStatus bool
}

type LabelName struct {
  Label       string
  Color       string
  Url         string
}
