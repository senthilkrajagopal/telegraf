// +build linux

package iptables

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Iptables is a telegraf plugin to gather packets and bytes throughput from Linux's iptables packet filter.
type Iptables struct {
	UseSudo bool
	UseLock bool
	Table   string
	Chains  []string
	lister  chainLister
}

// Description returns a short description of the plugin.
func (ipt *Iptables) Description() string {
	return "Gather packets and bytes throughput from iptables"
}

// SampleConfig returns sample configuration options.
func (ipt *Iptables) SampleConfig() string {
	return `
  ## iptables require root access on most systems.
  ## Setting 'use_sudo' to true will make use of sudo to run iptables.
  ## Users must configure sudo to allow telegraf user to run iptables with no password.
  ## iptables can be restricted to only list command "iptables -nvL"
  use_sudo = false
  ## Setting 'use_lock' to true runs iptables with the "-w" option.
  ## Adjust your sudo settings appropriately if using this option ("iptables -wnvl")
  use_lock = false
  ## defines the table to monitor:
  table = "filter"
  ## defines the chains to monitor:
  chains = [ "INPUT" ]
`
}

// Gather gathers iptables packets and bytes throughput from the configured tables and chains.
func (ipt *Iptables) Gather(acc telegraf.Accumulator) error {
	if ipt.Table == "" || len(ipt.Chains) == 0 {
		return nil
	}
	// best effort : we continue through the chains even if an error is encountered,
	// but we keep track of the last error.
	var err error
	for _, chain := range ipt.Chains {
		data, e := ipt.lister(ipt.Table, chain)
		if e != nil {
			err = e
			continue
		}
		e = ipt.parseAndGather(data, acc)
		if e != nil {
			err = e
			continue
		}
	}
	return err
}

func (ipt *Iptables) chainList(table, chain string) (string, error) {
	iptablePath, err := exec.LookPath("iptables")
	if err != nil {
		return "", err
	}
	var args []string
	name := iptablePath
	if ipt.UseSudo {
		name = "sudo"
		args = append(args, iptablePath)
	}
	iptablesBaseArgs := "-nvL"
	if ipt.UseLock {
		iptablesBaseArgs = "-wnvL"
	}
	args = append(args, iptablesBaseArgs, chain, "-t", table, "-x")
	c := exec.Command(name, args...)
	out, err := c.Output()
	return string(out), err
}

const measurement = "iptables"

var errParse = errors.New("Cannot parse iptables list information")
var chainNameRe = regexp.MustCompile(`^Chain\s+(\S+)`)
var fieldsHeaderRe = regexp.MustCompile(`^\s*pkts\s+bytes\s+`)
var valuesRe = regexp.MustCompile(`^\s*([0-9]+)\s+([0-9]+)\s+.*?(/\*\s(.*)\s\*/)?$`)

func (ipt *Iptables) parseAndGather(data string, acc telegraf.Accumulator) error {
	lines := strings.Split(data, "\n")
	if len(lines) < 3 {
		return nil
	}
	mchain := chainNameRe.FindStringSubmatch(lines[0])
	if mchain == nil {
		return errParse
	}
	if !fieldsHeaderRe.MatchString(lines[1]) {
		return errParse
	}
	for _, line := range lines[2:] {
		mv := valuesRe.FindAllStringSubmatch(line, -1)
		// best effort : if line does not match or rule is not commented forget about it
		if len(mv) == 0 || len(mv[0]) != 5 || mv[0][4] == "" {
			continue
		}
		tags := map[string]string{"table": ipt.Table, "chain": mchain[1], "ruleid": mv[0][4]}
		fields := make(map[string]interface{})
		// since parse error is already catched by the regexp,
		// we never enter ther error case here => no error check (but still need a test to cover the case)
		fields["pkts"], _ = strconv.ParseUint(mv[0][1], 10, 64)
		fields["bytes"], _ = strconv.ParseUint(mv[0][2], 10, 64)
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

type chainLister func(table, chain string) (string, error)

func init() {
	inputs.Add("iptables", func() telegraf.Input {
		ipt := new(Iptables)
		ipt.lister = ipt.chainList
		return ipt
	})
}
