# CONSTITUTION

## PRINCIPLES

- Keep `yp` focused on resolving local filesystem paths and copying them to the system clipboard; interactive browsing remains optional.
- Preserve explicit invocation intent: ordinary file and directory arguments copy directly, while no arguments or a trailing directory separator requests interactive browsing when `fzf` is available.
- Reject missing or inaccessible direct-copy inputs before writing to the clipboard.
- Keep execution local and dependency-light by favoring standard-library Go and delegating interactive selection and clipboard delivery to established local tools.

## CONSTRAINTS

<!-- TODO: define invariant rules that must never be violated -->

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Use native agent planning for research, clarification, design, and implementation planning.
- Before implementation, inspect code and repository memory; create or adopt `SPEC.md` when material rationale exists.
- After validation, curate feature rationale, project invariants, reusable practices, and domain knowledge into their scope-appropriate canonical documents.
- Allow a justified `not required` repository-memory decision when code and tests preserve the complete durable truth.
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less.
- Before delivery, audit the complete affected source/test scope; whole-project reconcile and scheduled maintenance audit the entire repository.
- Exclude documentation files, all `docs/**`, all `.kit/**`, `.kit.yaml`, ignored files, vendored dependencies, and proven generated files.
- Split oversized files by semantic responsibility while preserving stable public entry points and behavior; never use minification or arbitrary numbered chunks to claim compliance.
<!-- END KIT-MANAGED BASELINE RULES -->

## CHANGE CLASSIFICATION

<!-- all work falls into one of two tracks — classify before acting -->

### Repository-Memory Work

<!-- use when: consequential product rationale, architecture, cross-component behavior, or historical decisions must survive -->
<!-- workflow: native plan → create/adopt SPEC.md before code → implement → validate → curate repository memory -->
<!-- legacy staged documents: BRAINSTORM.md, legacy SPEC.md, PLAN.md, TASKS.md only when explicitly chosen -->

### Ad Hoc (Lightweight)

<!-- use when: bug fixes, security reviews, refactors, dependency updates, config changes, small refinements -->
<!-- workflow: understand → implement → verify -->
<!-- docs: update practical canonical docs when behavior changes -->
<!-- do not create feature SPEC.md solely for ceremony; report a justified not-required memory decision -->

### Ad Hoc with Existing Specs

<!-- if change touches code with existing spec docs: update them when rationale, behavior, requirements, or approach changes -->
<!-- leave them unchanged when code and tests communicate the complete durable truth -->

## NON-GOALS

- Provide shell integration or run as a resident or background process.
- Create, modify, move, or delete user-selected filesystem entries; missing paths are errors rather than prospective clipboard targets.
- Call remote services or APIs, or require accounts, credentials, or deployed infrastructure.

## DEFINITIONS

- **Direct copy**: The non-interactive flow that validates and resolves an existing file or directory argument before writing its path to the clipboard.
- **Browse request**: An invocation with no arguments or with one directory argument ending in a path separator; it uses the picker when `fzf` is available and otherwise copies the directory directly.
- **Picker**: The optional `fzf`-based workflow for interactively browsing a local directory and selecting the clipboard target.
- **Directory list**: A newline-separated clipboard payload containing immediate, non-hidden subdirectory paths.
