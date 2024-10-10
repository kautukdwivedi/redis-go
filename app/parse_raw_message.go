package main

import "unicode"

func parseRawMessage(msgBuf []byte) []*command {
	cmds := make([]*command, 0)

	var cmd *command

	for _, b := range msgBuf {
		bRune := rune(b)
		switch {
		case bRune == '*':
			cmd = &command{
				rawBytes: []byte{b},
				args:     make([]string, 0),
			}
			cmds = append(cmds, cmd)
		case unicode.IsDigit(bRune):
			if cmd == nil {
				continue
			}
			cmd.append(b)
		default:
			if cmd == nil {
				continue
			}
			n := len(cmds) - 1
			lastCmd := cmds[n]
			if len(lastCmd.rawBytes) == 1 && lastCmd.rawBytes[0] == '*' {
				cmds = cmds[:n]
				n = len(cmds) - 1
				cmd = cmds[n]
				cmd.append('*')
			}
			cmd.append(b)
		}
	}

	return cmds
}
