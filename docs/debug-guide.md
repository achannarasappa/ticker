# Debug Guide

## Visual Studio Code

### Setup

1. Install delve:
```sh
go install github.com/go-delve/delve/cmd/dlv@latest
```

2. Setup vscode:

In vscode, open command palette with <kbd>CTRL</kbd>+<kbd>SHIFT</kbd>+<kbd>P</kbd>, select `Open launch.json`, and add debug config block below to `launch.json`

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Attach",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "",
      "port": 2345,
      "host": "127.0.0.1",
      "showLog": true,
      "trace": "log",
      "logOutput": "rpc"
    }
  ]
}
```

### Debug

1. Start `ticker` with remote debugger using this command:
```sh
dlv --listen=:2345 --api-version 2 --log=true --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log-dest=./debugger.log --headless --accept-multiclient debug main.go
```

2. Open debug menu in vscode with <kbd>CTRL</kbd>+<kbd>SHIFT</kbd>+<kbd>ALT</kbd>+<kbd>D</kbd>

3. Select command `Attach` and press the adjacent `▷` button to start `ticker` with the debugger attached