# Go Todo CLI
[GitHub Project](https://github.com/users/ake3mio/projects/6/views/1?pane=info)


A command-line todo manager built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Cobra](https://github.com/spf13/cobra), and [Huh](https://github.com/charmbracelet/huh).

---

## Features

- **Interactive terminal UI** built with Bubble Tea & Lipgloss
- **Persistent storage** via Sqlite3
- **Keyboard navigation and shortcuts**
- **Autocompletion support**

---

## Design Notes

Each view (`add`, `list`) implements the [`tui.Model`](./internal/tui/model.go) interface:

The [`Runner`](./internal/tui/runner.go) orchestrates model lifecycles, manages cleanup, and handles transitions between commands.

- Uses **pointer receivers** for mutable Bubble Tea models
- Each model is self-contained and exposes a `Cleanup()` method for resource management
- **`sync.Once`** ensures idempotent cleanup
- **`tea.Sequence`** guarantees graceful shutdown after state transitions
- Separation of concerns between **UI**, **persistence**, and **control flow**

---

## Usage

### View and Manage tasks
```bash
todo
```

From the list view, you can:
- Toggle tasks as complete/incomplete
- Delete a task
- Switch back to the add view

**Shortcuts**
- `ctrl + h` - Toggle hiding completed tasks
- `ctrl + a` - Add a new task
- `delete/backspace` - Delete a selected task

---
### Add a Task
```bash
todo add
```

You’ll be prompted to enter:
- **Task name**
- **Due date (YYYY-MM-DD)**

Press **Enter** to save.  
After adding, you’ll be automatically taken to the task list view.

**Shortcuts**
- `ctrl + l` - Go to the list view

---

## Autocompletion

Enable Zsh autocompletion:
```bash
source <(todo completion zsh)
```

For Bash:
```bash
source <(todo completion bash)
```

---

## Persistence

Tasks are managed via the [`TodoRepository`](./internal/persistence/db.go) interface

You can plug in a custom storage backend - file, Postgres, Redis, etc.

---

## Building

This project uses [mage](https://github.com/magefile/mage) for build automation.


```bash
# Standard build
mage build

```


---

## Testing

Unit tests were written with [testify](https://github.com/stretchr/testify).  
Run tests with:

```bash
mage test
```

---

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - reactive TUI framework
- [Huh](https://github.com/charmbracelet/huh) - interactive input components
- [Cobra](https://github.com/spf13/cobra) - command-line framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - terminal styling
- [Testify](https://github.com/stretchr/testify) - testing & assertions

---

## Example

```bash
> todo
```
1. ![Enter Task Name.png](docs/Enter%20Task%20Name.png)
2. ![Enter Task Due Date.png](docs/Enter%20Task%20Due%20Date.png)
3. ![Task List.png](docs/Task%20List.png)