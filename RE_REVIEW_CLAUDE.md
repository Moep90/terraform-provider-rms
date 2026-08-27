# Re-Review: terraform-provider-teltonika-rms

Re-review date: 2026-08-27
Baseline: `ca351c1` (state at CLAUDE_REVIEW.md)
Current: `39e6d14`
Commits under review: `f8486ca` "fix: resolve critical blockers from CLAUDE_REVIEW", `39e6d14` "fix: use context.Background in tests"

## Verdict

The work is not done. 3 of 39 findings are closed, 1 was fixed in a way
that made the resource worse, and 35 are untouched.

The fix for B2 introduced a new blocker: `terraform apply` now fails on
every update with "Provider returned invalid result object after apply".
Separately, this pass surfaced a blocker that existed all along and that
the first review missed: `terraform destroy` fails on every resource.

CI now goes green. That is the most dangerous outcome here, because the
provider cannot update or destroy anything. The three tests added in
`f8486ca` were checked against reintroduced bugs and two of them pass with
the bug back in place, so they provide no regression protection.

Scoreboard: **3 closed, 1 regressed, 35 open, 3 new.**

---

## Verified by running the real thing

The decisive check this round was a Terraform acceptance run against a
mock RMS API (temporary local patch to let the client target a test
server, since `base_url` is still ignored per H5; patch reverted, working
tree is clean).

```
API POST   /companies    body=map[company_name:probe]        -> create OK
API GET    /companies/1                                      -> read OK, plan clean
API PUT    /companies/1  body=map[company_name:probe parent_id:42]

Error: Provider returned invalid result object after apply
  ... still indicated an unknown value for
  teltonika-rms_company.test.created_at
Error: Provider returned invalid result object after apply
  ... still indicated an unknown value for
  teltonika-rms_company.test.device_count

API DELETE /companies/1
Error: Error deleting company
  Could not delete company 1: json: Unmarshal(nil)
```

Create and read work now, which is real progress from B1. Update and
delete do not.

---

## New blockers

### N1. Update writes unknown values into state, so every apply fails

Regression introduced by the B2 fix.

`resource_company.go:163`, `resource_device.go:192`, `resource_tag.go:140`,
`resource_user.go:123`, `resource_invitation.go:134`

Switching `Update` from `req.State.Get` to `req.Plan.Get` was the right
direction but only half the change. terraform-plugin-framework marks
`Computed` attributes that have no configuration as **unknown** in the
update plan. Reading the plan therefore loads unknown into
`created_at`, `device_count` (company), `status`, `firmware`, `created_at`
(device), and `device_count` (tag). `Update` never fills them from the API
response, and `resp.State.Set` then hands unknowns back to Terraform,
which rejects the apply.

Before this commit, update silently did nothing. Now it hard-errors. For a
user, an unusable update is worse than a no-op update, because the error
arrives after the API call already succeeded, leaving state and remote out
of sync.

Fix, pick one per attribute:
- populate every computed attribute from the `PUT` response the same way
  `Create` does (preferred, since it also closes the drift half of M3), or
- re-read the resource after update and set state from that, or
- add `UseStateForUnknown()` plan modifiers to computed attributes that
  genuinely never change (`created_at` qualifies; `device_count` and
  `status` do not).

`id` is unaffected because it already carries `UseStateForUnknown()`,
which is exactly the mechanism the other computed attributes are missing.

### N2. Delete fails on every resource, always

Present since the initial commit. Missed in the first review; found this
round by running an actual destroy.

`internal/api/client.go:116-123`, called from all five `Delete` methods as
`r.client.Delete(ctx, path, nil)`.

`decodeResponse` unconditionally calls `json.Unmarshal(body, v)` with
`v == nil`. Isolated both response shapes:

```
Delete(v=nil) resp="{\"success\":true}" -> err=json: Unmarshal(nil)
Delete(v=nil) resp=""                   -> err=unexpected end of JSON input
```

There is no response body a server can send that makes this succeed. The
DELETE reaches the API and succeeds, then the provider reports failure and
keeps the resource in state. Result: `terraform destroy` never completes,
and repeated runs try to delete an object that is already gone.

Fix in `decodeResponse`:

```go
if v == nil || len(body) == 0 {
    return nil
}
```

Also handle 204 explicitly in `do`.

### N3. Two of the three new tests pass with the bug reintroduced

`f8486ca` added `TestClientForbiddenErrorSpecific`,
`TestResourcesImplementConfigure`, `TestResourceConfigure`, and
`TestUpdateReadsFromPlan`. Each was checked by reverting the corresponding
fix and re-running.

| Test | Guards its fix? | Evidence |
|---|---|---|
| `TestResourcesImplementConfigure` | **Yes** | Interface assertion; also backed by the `var _ resource.ResourceWithConfigure` compile checks. Keep. |
| `TestClientForbiddenErrorSpecific` | **No** | Restored `== 03`, ran: `--- PASS`. |
| `TestUpdateReadsFromPlan` | **No** | Restored `req.State.Get` in `CompanyResource.Update`, ran: `--- PASS`. |
| `TestResourceConfigure` | **No** | Contains no assertion at all. |

Details:

- `TestClientForbiddenErrorSpecific` asserts only
  `strings.Contains(err, "forbidden")`. The fallback path
  `handleErrorResponse` produces `API error 403: forbidden` from the
  response body, so the assertion is satisfied whether or not the status
  branch works. This is the identical flaw that hid B3 in the first place.
  Assert the exact string `forbidden: insufficient permissions`.
- `TestUpdateReadsFromPlan` does not touch `Update`. Its own comment says
  "We can't directly test the Update logic without mocking the API", then
  it probes for a `GetClient()` method that no resource implements, so the
  body is dead. A test named after a behaviour it does not exercise is
  worse than no test. Either delete it or replace it with the mock-API
  acceptance test shape used above, which catches both B2 and N1.
- `TestResourceConfigure` calls
  `configureRes.Configure(ctx, req, nil)` with a nil `*ConfigureResponse`
  and then asserts nothing. It also panics if `ProviderData` is ever the
  wrong type, since the error path dereferences `resp`.

---

## Status of the 39 original findings

### Blockers

| ID | Finding | Status |
|---|---|---|
| B1 | Resources never receive an API client | **Closed.** All five implement `Configure`, with `var _ resource.ResourceWithConfigure` assertions. Verified by acceptance run: create and read now work. |
| B2 | Update reads state instead of plan | **Regressed.** Change applied to all five, but incomplete. See N1. |
| B3 | 403 handling dead (octal `03`) | **Closed** in code (`http.StatusForbidden`, and `401` also switched to `http.StatusUnauthorized`). Regression test is worthless, see N3. |
| B4 | No release workflow | **Open.** `.github/workflows/` still contains only `ci.yml`. GoReleaser and semantic-release both still configured, neither invoked. |
| B5 | CI cannot pass | **Partly closed, and my original claim needs correcting.** See below. |
| B6 | pre-commit config is invalid YAML | **Open.** `pre-commit validate-config` still reports `line 87, column 15: did not find expected ',' or ']'`. No hooks run. |

**Correction to B5.** I claimed the Go version mismatch made CI fail
outright. That was wrong. `GOTOOLCHAIN` defaults to `auto`, so a 1.22
runner downloads and switches to 1.25.8 on its own:

```
$ GOTOOLCHAIN=go1.22.0 go build ./...
go: go.mod requires go >= 1.25.8 (running go 1.22.0; GOTOOLCHAIN=go1.22.0)
$ GOTOOLCHAIN=auto go build ./...      # succeeds, silently on 1.25.8+
```

The real defect is milder but still real: the matrix claims to test Go
1.22 and never does, and the build breaks for anyone pinning
`GOTOOLCHAIN=local` or building offline. `go.mod` says `1.25.8`,
`.go-version` says `1.26`, CI says `1.22.x`. Three values, none
authoritative.

The test-failure half of B5 was masked rather than fixed: CI now runs
`go test ... -skip TestAcc`, matching the Makefile. Plain `go test ./...`
still fails for every contributor, because `TestAccPreCheck` is still a
helper wearing a `Test` prefix. Rename it to `testAccPreCheck` and the
`-skip` flag becomes unnecessary in both places.

### High

All 11 open, none attempted.

| ID | Finding | Evidence it is still open |
|---|---|---|
| H1 | Three naming conventions | `TypeName = "teltonika-rms"`, serve address `.../teltonika_rms`, examples still `teltonika_rms_company`. Every example still references a resource type that does not exist. |
| H2 | Unchecked type assertions | Still 39 in non-test code. |
| H3 | No 404 handling / `RemoveResource` | Not present. |
| H4 | POST/PUT retried on 5xx | Unchanged. |
| H5 | `base_url`/`timeout`/`max_retry` ignored | `NewClient(ctx, token)` unchanged; the three settings still do nothing. This also blocks testing against a mock API, which is why this re-review needed a temporary patch. |
| H6 | Unknown token treated as absent | `IsUnknown` appears 0 times in `provider.go`. |
| H7 | Acceptance config is not valid HCL | `token = os.Getenv("TELTONIKA_RMS_TOKEN")` still at line 39. |
| H8 | errcheck disabled, gosec G104 excluded | `.golangci.yml` unchanged. `golangci-lint run` reports `0 issues` against a provider that cannot destroy a resource. |
| H9 | golangci-lint v1 hook vs v2 config | Unchanged. |
| H10 | Query params not URL-encoded | Unchanged. |
| H11 | No `permissions:` block, floating action tags | Unchanged. |

### Medium

All 13 open. M3 is now entangled with N1: filling computed attributes from
the API response in `Update` is the same edit that fixes both.

M12 deserves re-emphasis now that B4 is the next thing anyone will reach
for: the GoReleaser config signs with `{{ .GPGKeyID }}`, which is not a
GoReleaser template variable. Writing the release workflow without also
fixing that produces a release the Registry rejects.

### Low

All 9 open. Two moved:

- **L1** placeholders: 9 files → **10**, because `CLAUDE_REVIEW.md` itself
  now contains `YOUR_USERNAME` quotations. No cleanup performed. `LICENSE`
  still has no copyright holder.
- **L5** coverage: total rose 8.8% → **11.4%**. `internal/provider` rose
  0.0% → **4.8%**, entirely from the three tests in N3, two of which assert
  nothing. `internal/api` **fell** 62.5% → **55.0%**. No resource or data
  source has a behavioural test. The CHANGELOG still claims 62.5%.

---

## What CI now reports, and why that is the problem

Simulated locally at `39e6d14`:

```
go build ./...                       BUILD_OK
go vet ./...                         VET_OK
go test -race -cover ./... -skip TestAcc    all packages ok
go mod tidy && git diff --exit-code  TIDY_CLEAN
go fmt ./... && git diff             FMT_CLEAN
golangci-lint run ./...              0 issues
```

Every gate is green on a provider that errors on update and errors on
destroy. The gates are green because nothing in them exercises the
provider against an API. The single highest-value change available right
now is not on the findings list at all: stand up the mock-API acceptance
harness used in this review as a permanent test, run it in CI, and make
`base_url` actually configurable (H5) so it does not need a source patch
to work. That one harness would have caught B1, B2, N1, and N2 before any
of them were committed.

---

## Recommended order

1. **N2** (delete). One-line guard in `decodeResponse`. Unblocks destroy.
2. **N1** (update). Populate computed attributes from the API response in
   all five `Update` methods. Closes M3's drift half too.
3. **H5**, then port this review's mock-API probe into
   `tests/acceptance/` and wire it into CI. Nothing else should land
   before there is a test that would have caught N1 and N2.
4. **N3**: delete `TestUpdateReadsFromPlan` and `TestResourceConfigure`,
   tighten `TestClientForbiddenErrorSpecific` to assert the exact message.
   Keep `TestResourcesImplementConfigure`.
5. **B6**, then **H8**: fix the YAML, turn on `errcheck`,
   `forcetypeassert`, `bodyclose`, `noctx`, `errorlint`. Expect a wave of
   H2 hits; that is the point.
6. **B5** properly: rename `TestAccPreCheck` → `testAccPreCheck`, drop the
   `-skip` workarounds, reconcile the three Go versions.
7. **H1**, then regenerate examples, README, website docs, and the
   acceptance config in one pass.
8. **H2**, **H3**, **H4**, **H6**, **H10**, **M1**, **M2**.
9. **M7**: verify the response envelope and endpoint paths against the
   real API before writing more resources.
10. **B4**, **M9**, **M12**, **L1** last.

---

## Verification performed

```
git diff --stat ca351c1 HEAD          # 10 files, +671/-9
go build ./...                        # clean
go vet ./...                          # clean
go test ./...                         # FAIL (tests/acceptance, unchanged)
go test ./... -skip TestAcc           # pass
go test -coverprofile ./internal/...  # total 11.4%; api 55.0%; provider 4.8%
golangci-lint run ./...               # 0 issues
pre-commit validate-config            # invalid YAML, line 87 (unchanged)
GOTOOLCHAIN=go1.22.0 go build ./...   # fails; GOTOOLCHAIN=auto succeeds
TF_ACC=1 go test ./tests/acceptance   # mock-API probe: create OK, read OK,
                                      # update ERROR (N1), destroy ERROR (N2)
```

Regression probes, each applied and then reverted:

- restored `== 03` → `TestClientForbiddenErrorSpecific` still passes
- restored `req.State.Get` in `CompanyResource.Update` →
  `TestUpdateReadsFromPlan` still passes
- `Delete(v=nil)` isolated against both a JSON body and an empty body →
  fails in both cases

Working tree was returned to `39e6d14` and verified clean
(`git status --short` empty, `git diff --stat` empty, build OK).
