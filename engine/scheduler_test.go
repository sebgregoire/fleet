package engine

import (
	"reflect"
	"sort"
	"testing"

	"github.com/coreos/fleet/agent"
	"github.com/coreos/fleet/job"
	"github.com/coreos/fleet/machine"
)

func TestSchedulerDecisions(t *testing.T) {
	tests := []struct {
		clust *clusterState
		job   *job.Job
		dec   *decision
	}{
		// no machines to receive job
		{
			clust: newClusterState([]job.Job{}, []machine.MachineState{}),
			job:   &job.Job{Name: "foo.service"},
			dec:   nil,
		},

		// multiple machines, pick the first one
		{
			clust: newClusterState([]job.Job{}, []machine.MachineState{machine.MachineState{ID: "XXX"}, machine.MachineState{ID: "YYY"}}),
			job:   &job.Job{Name: "foo.service"},
			dec: &decision{
				machineID: "XXX",
			},
		},
	}

	for i, tt := range tests {
		sched := &leastLoadedScheduler{}
		dec, err := sched.Decide(tt.clust, tt.job)

		if err != nil && tt.dec != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
			continue
		} else if err == nil && tt.dec == nil {
			t.Errorf("case %d: expected error", i)
			continue
		}

		if !reflect.DeepEqual(tt.dec, dec) {
			t.Errorf("case %d: expected decision %#v, got %#v", i, tt.dec, dec)
		}
	}
}

func TestAgentStateSorting(t *testing.T) {
	tests := []struct {
		in  []*agent.AgentState
		out []*agent.AgentState
	}{
		{
			in:  []*agent.AgentState{},
			out: []*agent.AgentState{},
		},

		// sort by number of jobs scheduled to agent
		{
			in: []*agent.AgentState{
				&agent.AgentState{
					MState: &machine.MachineState{ID: "A"},
					Jobs: map[string]*job.Job{
						"1.service": &job.Job{},
						"2.service": &job.Job{},
						"3.service": &job.Job{},
						"4.service": &job.Job{},
						"5.service": &job.Job{},
					},
				},
				&agent.AgentState{
					MState: &machine.MachineState{ID: "B"},
					Jobs: map[string]*job.Job{
						"6.service": &job.Job{},
						"7.service": &job.Job{},
					},
				},
			},
			out: []*agent.AgentState{
				&agent.AgentState{
					MState: &machine.MachineState{ID: "B"},
					Jobs: map[string]*job.Job{
						"6.service": &job.Job{},
						"7.service": &job.Job{},
					},
				},
				&agent.AgentState{
					MState: &machine.MachineState{ID: "A"},
					Jobs: map[string]*job.Job{
						"1.service": &job.Job{},
						"2.service": &job.Job{},
						"3.service": &job.Job{},
						"4.service": &job.Job{},
						"5.service": &job.Job{},
					},
				},
			},
		},

		// fall back to sorting alphabetically by machine ID when # jobs is equal
		{
			in: []*agent.AgentState{
				&agent.AgentState{
					MState: &machine.MachineState{ID: "B"},
					Jobs: map[string]*job.Job{
						"1.service": &job.Job{},
						"2.service": &job.Job{},
					},
				},
				&agent.AgentState{
					MState: &machine.MachineState{ID: "A"},
					Jobs: map[string]*job.Job{
						"3.service": &job.Job{},
						"4.service": &job.Job{},
					},
				},
			},
			out: []*agent.AgentState{
				&agent.AgentState{
					MState: &machine.MachineState{ID: "A"},
					Jobs: map[string]*job.Job{
						"3.service": &job.Job{},
						"4.service": &job.Job{},
					},
				},
				&agent.AgentState{
					MState: &machine.MachineState{ID: "B"},
					Jobs: map[string]*job.Job{
						"1.service": &job.Job{},
						"2.service": &job.Job{},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		sortable := make(sortableAgentStates, len(tt.in))
		for i, ms := range tt.in {
			ms := ms
			sortable[i] = ms
		}

		sort.Sort(sortable)
		sorted := []*agent.AgentState(sortable)

		if !reflect.DeepEqual(tt.out, sorted) {
			t.Errorf("case %d: unexpected output", i)
			for ii, ms := range tt.out {
				t.Logf("case %d: tt.out[%d] = %#v", i, ii, *ms)
			}
			for ii, ms := range sorted {
				t.Logf("case %d: sorted[%d] = %#v", i, ii, *ms)
			}
		}
	}
}
