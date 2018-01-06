package cli

import "fmt"

type cmdEngine struct {
	Handlers map[string]CmdHandler
}

type CmdHandler func() error

var (
	engine = new(cmdEngine)
)

func RegCmdHandler(cmd string, handler CmdHandler) {
	if engine.Handlers == nil {
		engine.Handlers = make(map[string]CmdHandler)
	}
	if _, ok := engine.Handlers[cmd]; ok {
		fmt.Println("[WARNING] Re-registration of", cmd, "command handler")
	}
	engine.Handlers[cmd] = handler
}

func RunCmd(cmd string) error {
	if engine.Handlers == nil {
		return fmt.Errorf("Processing engine commands not ready")
	}
	handler, ok := engine.Handlers[cmd]
	if !ok {
		return fmt.Errorf("Command \"%s\" not found", cmd)
	}
	return handler()
}
