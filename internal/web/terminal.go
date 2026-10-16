package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/creack/pty"

	"github.com/Sheila724/incidente-zero/internal/lab"
)

// Mensagens de controle vindas do navegador (frames de texto).
// Frames binários são digitação e vão direto para o terminal.
type termMsg struct {
	Type string `json:"type"` // "resize" | "input"
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	Data string `json:"data"`
}

func (s *Server) terminal(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.findRunning(w, r)
	if !ok {
		return
	}

	// Accept, por padrão, só aceita conexões da mesma origem.
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.CloseNow()

	cmd := lab.ShellCommand(sc)
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 30, Cols: 100})
	if err != nil {
		c.Close(websocket.StatusInternalError, "não foi possível abrir o terminal")
		return
	}
	defer func() {
		_ = ptmx.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// terminal → navegador
	go func() {
		defer cancel()
		buf := make([]byte, 32*1024)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				if werr := c.Write(ctx, websocket.MessageBinary, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				c.Close(websocket.StatusNormalClosure, "sessão encerrada")
				return
			}
		}
	}()

	// navegador → terminal
	for {
		typ, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		if typ == websocket.MessageBinary {
			_, _ = ptmx.Write(data)
			continue
		}
		var m termMsg
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		switch m.Type {
		case "resize":
			if m.Cols > 0 && m.Rows > 0 {
				_ = pty.Setsize(ptmx, &pty.Winsize{Rows: m.Rows, Cols: m.Cols})
			}
		case "input":
			_, _ = ptmx.Write([]byte(m.Data))
		}
	}
}
