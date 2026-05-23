# AGENTS.md - Technical Notes for UPF Tester

This file contains project-specific guidance for AI coding agents and future development sessions. Treat it as the operational companion to `README.md`: the README explains the project to humans; this file records implementation structure, conventions, and common pitfalls for agents modifying the codebase.

## Project Overview

UPF Tester is a Go-based UPF (User Plane Function) test tool. It exercises both the PFCP control plane and the GTP-U/ICMP data plane through configurable YAML test flows.

Core goals:

- Drive PFCP session lifecycle operations against a UPF.
- Establish, modify, and delete sessions through YAML-driven test cases.
- Associate session context with data-plane tests using SEID, TEID, and UE IP.
- Send and receive GTP-U encapsulated traffic for connectivity validation.
- Keep the tool simple enough to run directly from the command line against lab UPF instances such as free5GC UPF.

## Repository Structure

### Command Entry Point

**`cmd/`**

- Contains the executable entry point.
- Build from this directory unless the project layout changes.
- Current README build pattern:

```bash
cd cmd
go build -o upf-tester main.go
```

Run pattern:

```bash
cd cmd
./upf-tester
```

When modifying startup behavior, confirm the relative paths used for config and test case loading still work when invoked from `cmd/`.

### Configuration

**`config/config.yaml`**

Holds lab network parameters and resource allocation defaults.

Important fields:

- `basic.localN4Ip`: local SMF-side N4 address.
- `basic.upfN4Ip`: UPF N4 address.
- `dataPlane.gnbIp`: simulated gNB address.
- `dataPlane.n3Ip`: UPF N3 address.
- `dataPlane.n6Ip`: UPF N6 address.
- `dataPlane.dnIp`: data network address.
- `resources.queueSize`: packet/event queue capacity.
- `resources.startUeIp`: first UE IP to allocate.
- `resources.startSeId`: initial SEID.
- `resources.startTeTd`: initial TEID. Preserve the existing key spelling unless the config loader is also migrated.

Do not hard-code lab IPs in Go code. Prefer config-driven behavior.

### Test Cases

**`testcases/`**

Contains YAML-based test flow definitions and per-step message/test configuration.

Typical flow:

1. Send `session_establishment_request`.
2. Receive `session_establishment_response`.
3. Run `data_plane_test` with `icmp` action.
4. Send `session_modification_request`.
5. Receive `session_modification_response`.
6. Send `session_deletion_request`.
7. Receive `session_deletion_response`.

Agents adding new test steps should update:

- The test case YAML schema/examples.
- The test case executor switch/dispatch logic.
- README documentation if the new step is user-facing.
- This file if the change affects agent behavior or project conventions.

### Control Plane Layer

**`internal/handler/`**

Expected responsibilities:

- PFCP message dispatch.
- Association setup/release handling.
- Test case execution.
- Session context tracking.

Key files referenced by project documentation:

- `pfcphandler.go`: PFCP message dispatcher.
- `assochandler.go`: Association handling.
- `testcasehandler.go`: test case executor.
- `session_context.go`: session context management.

When modifying PFCP flows, preserve explicit handling of session lifecycle state. Session establishment must produce enough context for later modification, deletion, and data-plane tests.

### PFCP Encoding Layer

**`encoding/pfcp/`**

Expected responsibilities:

- PFCP request encoding.
- Shared PFCP type definitions.
- Session establishment, modification, and deletion message construction.

Key files referenced by project documentation:

- `establishmentrequest.go`
- `modificationrequest.go`
- `deletionrequest.go`
- `types.go`

When adding or changing PFCP IEs, keep message construction deterministic and readable. Prefer small helper functions for repeated IE assembly.

### Data Plane Layer

**`internal/dataplane/`**

Expected responsibilities:

- Data-plane test orchestration.
- GTP-U packet encapsulation.
- ICMP message construction.
- Packet send/receive and result accounting.

Key files referenced by project documentation:

- `test.go`: data-plane test framework.
- `sender.go`: packet sender.
- `receiver.go`: packet receiver.
- `gtp.go`: GTP-U encapsulation.
- `icmp.go`: ICMP construction.

Data-plane tests must use the active session context instead of independently inventing TEID or UE IP values.

### Utility Layer

**`internal/util/`**

Expected responsibilities:

- SEID allocation.
- TEID allocation.
- Sequence number management.

Key files referenced by project documentation:

- `seid.go`
- `seqnumber.go`
- `teid.go`

Allocation utilities should remain concurrency-safe if test execution becomes parallelized.

## Development Conventions

### Language and Style

- Primary language: Go.
- Prefer idiomatic Go over framework-heavy abstractions.
- Keep protocol-specific code explicit; avoid hiding PFCP/GTP-U details behind overly generic abstractions.
- Keep logs specific enough to debug packet/session behavior in lab environments.
- Use clear error wrapping so failures identify the failing step, message type, and relevant SEID/TEID where possible.

### Config and Paths

- Avoid absolute paths such as `/localdisk/upf-tester` in code.
- Support execution from documented working directories.
- When changing path handling, test from both repository root and `cmd/` if feasible.

### YAML Compatibility

Be conservative with YAML schema changes.

If a schema change is unavoidable:

1. Keep backward compatibility where possible.
2. Update example YAML files.
3. Update README sections describing test step configuration.
4. Add validation errors that tell users which field is missing or invalid.

### Session Context Rules

Session context is central to the project.

Preserve these invariants:

- Session establishment creates or updates the current session context.
- Session modification targets an established session.
- Session deletion cleans up the matching session context.
- Data-plane tests read TEID, SEID, and UE IP from the current session context.
- Logs should identify both SMF-side and UPF-side SEIDs when available.

### Error Handling Philosophy

- Fail fast for malformed configuration or invalid test case YAML.
- Continue only when continuing cannot corrupt session state or produce misleading test results.
- Do not silently ignore packet send/receive errors.
- For expected network timeouts, report them as test results rather than generic crashes.
- Distinguish protocol failure, config failure, socket/network failure, and test assertion failure.

### Logging

Useful logs should include:

- Test step number and type.
- PFCP message type.
- SEID and TEID values.
- UE IP.
- Remote/local endpoint addresses.
- Packet counts, loss rate, and timeout reason for data-plane tests.

Avoid excessive per-packet logs unless behind a debug flag.

## Common Gotchas

1. **Working directory assumptions**
   The README currently builds and runs from `cmd/`. Relative config paths may depend on this. Check path behavior before changing startup code.

2. **N4/N3/N6 address confusion**
   Keep control-plane N4 addresses separate from data-plane N3/N6 addresses. Do not reuse config fields casually.

3. **SEID directionality**
   Be explicit about local/SMF SEID versus UPF SEID. Bugs here can make deletion or modification target the wrong session.

4. **TEID lifecycle**
   TEID allocation must match the session used by GTP-U encapsulation. Do not generate a fresh TEID inside a data-plane test unless the test case explicitly requires it.

5. **YAML step dispatch**
   Adding a new `type` or `action` usually requires both YAML examples and Go dispatch logic updates.

6. **ICMP test interpretation**
   Packet loss may be caused by lab routing, UPF forwarding, GTP-U encapsulation, firewall rules, or wrong TEID. Do not assume one cause without evidence from logs and packet captures.

7. **Resource key spelling**
   The README documents `startTeTd`. Before renaming it to `startTeId` or `startTeid`, verify the current config loader and existing YAML files.

## Suggested Validation Before Committing

Run the following where applicable:

```bash
go fmt ./...
go test ./...
```

Build the CLI:

```bash
cd cmd
go build -o upf-tester main.go
```

If a change affects runtime behavior, validate with a representative YAML flow against a reachable UPF lab environment.

## Future Enhancement Ideas

These are reasonable directions, but do not implement them opportunistically unless requested:

- Add explicit config validation with actionable error messages.
- Add dry-run mode for validating YAML flows without sending packets.
- Add pcap capture or packet dump hooks for data-plane debugging.
- Add structured JSON output for CI/test automation.
- Add concurrent multi-session test flows.
- Add more data-plane protocols beyond ICMP.
- Add table-driven unit tests for PFCP IE construction.
- Add integration test profiles for free5GC UPF.

## Agent Workflow

When making changes as an AI coding agent:

1. Inspect the current file before editing it.
2. Prefer minimal patches over broad rewrites.
3. Preserve documented CLI behavior unless the task explicitly changes it.
4. Update docs and sample YAML when behavior changes.
5. Run formatting and tests when tool access permits.
6. Report exactly what changed, what was not tested, and any assumptions made.

## Data Flow Summary

```text
YAML test case
    ↓
Test case executor
    ↓
PFCP control-plane operation
    ↓
Session context allocation/tracking
    ↓
Optional data-plane test
    ↓
GTP-U/ICMP send-receive loop
    ↓
Statistics and logs
    ↓
Session modification/deletion cleanup
```

Keep this control-plane/data-plane coupling clear when adding features.
