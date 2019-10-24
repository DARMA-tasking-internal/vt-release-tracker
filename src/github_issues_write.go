package main

import (
  "fmt"
  "os"
  "strconv"
  "net/http"
  "io/ioutil"
  "bytes"
  //"encoding/json"
)

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
