# Tasks

## TODO

- [x] Create ticket VM-017-MERGE-UI-VM and scaffold docs
- [x] Inventory vm-system and vm-system-ui architecture/build/runtime surfaces
- [x] Compare merge approaches (submodule, copy, filter-repo, subtree)
- [x] Produce in-depth design + analysis document with recommended strategy
- [x] Maintain detailed implementation diary while working
- [x] Upload design + diary bundle to reMarkable
- [x] Verify cloud listing for uploaded analysis bundle

## Implementation Checklist

- [x] Create implementation branch and pre-merge safeguards
- [x] Import vm-system-ui into vm-system as `ui/` with preserved git history
- [x] Add Go web static serving package (`internal/web`) with disk + embed modes
- [x] Wire SPA/static handler into daemon serve path without shadowing `/api/v1`
- [x] Add generator bridge to build/copy frontend assets into Go embed path
- [x] Add developer commands (`Makefile`) for backend/frontend dev and embed build
- [x] Validate backend tests and frontend checks/build in merged repository
- [ ] Update VM-017 design/index/changelog/diary with implementation outcomes
- [ ] Upload implementation update bundle to reMarkable
- [x] Remove __manus__ and other imported Manus/debug junk from ui
