package redisutils

import (
	"errors"
	"fmt"
)

type multiCmds struct {
	cmds [][]interface{}
}

func EmptyMultiCmd() *multiCmds {
	cmds := &multiCmds{}
	cmds.cmds = make([][]interface{}, 0)
	return cmds
}

func MultiCmd(cmdName string, args ...interface{}) *multiCmds {
	cmds := &multiCmds{}
	cmds.cmds = make([][]interface{}, 0)
	cmd := make([]interface{}, 0)
	cmd = append(cmd, cmdName)
	for _, arg := range args {
		cmd = append(cmd, arg)
	}

	cmds.cmds = append(cmds.cmds, cmd)
	return cmds
}

// 追加指令
func (multi *multiCmds) Append(cmdName string, args ...interface{}) *multiCmds {
	cmd := make([]interface{}, 0)
	cmd = append(cmd, cmdName)
	for _, arg := range args {
		cmd = append(cmd, arg)
	}

	multi.cmds = append(multi.cmds, cmd)
	return multi
}

// 批量执行
func (multi *multiCmds) Exec(dbIndex int) error {
	if dbIndex < 0 || dbIndex > 15 {
		return errors.New("dbIndex must in [0,15]")
	}

	if multi == nil || len(multi.cmds) < 1 {
		return errors.New("invalid commands")
	}

	conn := getConn0(uint8(dbIndex))
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	for i, cmd := range multi.cmds {
		if len(cmd) < 2 {
			return errors.New(fmt.Sprintf("invalid command at position %d", i))
		}

		key := cmd[0]
		cmdName, ok := key.(string)
		if !ok {
			return errors.New(fmt.Sprintf("command name not string at position %d", i))
		}

		args := cmd[1:]

		if err := conn.Send(cmdName, args...); err != nil {
			return err
		}
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}
