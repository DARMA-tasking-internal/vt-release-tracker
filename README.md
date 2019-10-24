
This script analyzes a git repository (like *vt*) to test if feature branches
created from issues with Github labels properly merged. It uses both `git branch--merged`
and greps from `git log` to find matching commits for a given branch. It uses the Github
API to pull issues and PRs, along with their associated labels, to check the merge state
automatically.

## Building
- Get `go`
- `cd src`
- `go build -o vt-release-tracker`
- `./vt-release-tracker <branch> <github-label>...`
  - Example: `./vt-release-tracker 1.0.0 1.0.0-beta.1 1.0.0-beta.2`

## Example Output

```
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
                                                          RESULTS FROM ANALYSIS OF BRANCH "1.0.0-beta.2"
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
| Issue  | PR     | State                | Issue State     | Branch                                        | Matching labels                                    |
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
| 450    | 454    | Merged Correctly     | closed          | 450-cmd-line-ordering-error                   | 1.0.0-beta.1                                       |
| 460    | 461    | Merged Correctly     | closed          | 460-argv-fix                                  | 1.0.0-beta.1                                       |
| 462    | 463    | Merged Correctly     | closed          | 462-lb-test-init-crash                        | 1.0.0-beta.1                                       |
| 448    |        | Merged Correctly     | closed          | 448-enum-serialization                        | 1.0.0-beta.1                                       |
| 507    | 508    | Merged Correctly     | closed          | 507-hierlb-race-condition                     | 1.0.0-beta.2                                       |
| 427    | 429    | Merged Correctly     | closed          | 427-lb-base-class                             | 1.0.0-beta.1                                       |
| 457    | 468    | Merged Correctly     | closed          | 457-cleanup-gcc-warnings                      | 1.0.0-beta.1                                       |
| 459    | 495    | Merged Correctly     | closed          | 459-fix-template-syntax                       | 1.0.0-beta.1                                       |
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
| 382    | 383    | Merged Incorrectly   | closed          | 382-compiler-version                          |                                                    |
| 446    | 447    | Merged Incorrectly   | closed          | 446-missing-hash-function                     |                                                    |
| 437    | 438    | Merged Incorrectly   | closed          | 437-query-index-context                       |                                                    |
| 408    | 409    | Merged Incorrectly   | closed          | 408-test-term-chain-integration-test          |                                                    |
| 378    |        | Merged Incorrectly   | closed          | 378-serialize-epoch-lost                      |                                                    |
| 399    | 400    | Merged Incorrectly   | closed          | 399-intel-static-template-variable            |                                                    |
| 417    |        | Merged Incorrectly   | closed          | 417-transfer-changes-EMPIRE-on-master         |                                                    |
| 484    | 485    | Merged Incorrectly   | closed          | 484-insertable-collection-use-after-destroy   |                                                    |
| 244    | 381    | Merged Incorrectly   | open            | 244-trace-user-event                          |                                                    |
| 419    | 422    | Merged Incorrectly   | closed          | 419-lb-fpe-crash                              |                                                    |
| 458    |        | Merged Incorrectly   | closed          | 458-compile-clang-8.0                         |                                                    |
| 410    | 425    | Merged Incorrectly   | open            | 410-integral-set-only                         |                                                    |
| 272    |        | Merged Incorrectly   | closed          | 272-example                                   |                                                    |
| 439    | 497    | Merged Incorrectly   | closed          | 439-fix-termination-hang                      |                                                    |
| 246    |        | Merged Incorrectly   | closed          | 246-callable                                  |                                                    |
| 434    | 436    | Merged Incorrectly   | closed          | 434-vt-print-disabled                         |                                                    |
| 388    | 389    | Merged Incorrectly   | closed          | 388-jacobi2d-abs-compile-error                |                                                    |
| 420    | 421    | Merged Incorrectly   | closed          | 420-no-terminate-startup                      |                                                    |
| 384    | 392    | Merged Incorrectly   | closed          | 384-nvcc-more-fixes                           |                                                    |
| 411    | 412    | Merged Incorrectly   | closed          | 411-apple-clang-compile-bug                   |                                                    |
| 416    | 435    | Merged Incorrectly   | closed          | 416-fix-non-standard-cmake                    |                                                    |
| 341    | 398    | Merged Incorrectly   | closed          | 341-terminate-_exit                           |                                                    |
| 404    | 405    | Merged Incorrectly   | closed          | 404-epoch-term-objgroup                       |                                                    |
| 452    | 453    | Merged Incorrectly   | closed          | 452-collection-context-allow-test             |                                                    |
| 441    | 442    | Merged Incorrectly   | closed          | 441-bug-in-lb                                 |                                                    |
| 490    | 491    | Merged Incorrectly   | closed          | 490-skewness-sigfpe-crash                     |                                                    |
| 386    | 387    | Merged Incorrectly   | closed          | 386-lb-bug-elm-id                             |                                                    |
| 406    | 407    | Merged Incorrectly   | closed          | 406-unify-ds-term                             |                                                    |
| 396    | 397    | Merged Incorrectly   | closed          | 396-trace-msg-size-bug                        |                                                    |
| 443    | 445    | Merged Incorrectly   | closed          | 443-license-update                            |                                                    |
| 489    | 493    | Merged Incorrectly   | closed          | 489-remove-global-cli11-include               |                                                    |
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
| 482    | 483    | Unmerged Incorrectly | closed          | 482-arg-array-overrun                         | 1.0.0-beta.1                                       |
| 428    | 487    | Unmerged Incorrectly | closed          | 428-auto_registry_warning                     | 1.0.0-beta.1                                       |
| 455    | 488    | Unmerged Incorrectly | closed          | 455-trace-dir-specification                   | 1.0.0-beta.1                                       |
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
```