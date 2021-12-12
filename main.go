package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

func getConnectionByPWD(path string, p *process.Process) (*nvim.Nvim, error) {
	if p == nil {
		return nil, errors.Errorf("%q", "no process provided")
	}

	cons, err := p.Connections()
	if err != nil {
		return nil, errors.Errorf("%q: %w", "could not get connection", err)
	}

	for _, c := range cons {
		v, err := nvim.Dial(c.Laddr.IP)
		if err != nil {
			if v != nil {
				v.Close()
			}
			continue
		}

		pwd, err := v.CommandOutput("pwd")
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(path, pwd) {
			return v, nil
		}
		v.Close()
	}

	return nil, nil
}

func main() {
	var v *nvim.Nvim

	if len(os.Args) < 2 {
		return
	}

	filePath := os.Args[1]

	ps, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range ps {
		name, err := p.Name()
		if err != nil {
			log.Fatal(err)
			continue
		}

		if name == "nvim" {
			con, err := getConnectionByPWD(filePath, p)
			if err != nil {
				continue
			}

			if con != nil {
				v = con
				defer v.Close()
				break
			}
		}
	}

	if v == nil {
		return
	}

	err = v.Command(fmt.Sprintf("e %s", filePath))
	if err != nil {
		log.Fatal(err)
	}
}
