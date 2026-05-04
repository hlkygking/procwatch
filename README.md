# procwatch

Lightweight process supervisor with structured log output and restart policies.

---

## Installation

```bash
go install github.com/yourusername/procwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/procwatch.git && cd procwatch && go build -o procwatch .
```

---

## Usage

Define your processes in a `procwatch.yaml` config file:

```yaml
processes:
  web:
    command: "./bin/server"
    restart: on-failure
    max_restarts: 5
  worker:
    command: "python worker.py"
    restart: always
```

Then run:

```bash
procwatch start -c procwatch.yaml
```

procwatch will supervise all defined processes, automatically restarting them according to the configured policy, and emit structured JSON logs to stdout:

```json
{"time":"2024-01-15T10:23:01Z","level":"info","process":"web","pid":12345,"msg":"process started"}
{"time":"2024-01-15T10:23:05Z","level":"warn","process":"worker","pid":12346,"msg":"process exited, restarting"}
```

### Restart Policies

| Policy | Behavior |
|---|---|
| `always` | Restart regardless of exit code |
| `on-failure` | Restart only on non-zero exit |
| `never` | Do not restart |

---

## License

MIT © [yourusername](https://github.com/yourusername)