package team4

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

func makeHistory(clientVars map[shared.ClientID][]rules.VariableFieldName, lieCounts map[shared.ClientID]int) (history []shared.Accountability, expected map[shared.ClientID]judgeHistoryInfo) {
	history = []shared.Accountability{}
	expected = map[shared.ClientID]judgeHistoryInfo{}
	for client, vars := range clientVars {
		pairs := []rules.VariableValuePair{}
		for _, v := range vars {
			newVal := rand.ExpFloat64()
			newPair := makeSingleVar(v, newVal)
			newAcc := shared.Accountability{
				ClientID: client,
				Pairs:    []rules.VariableValuePair{newPair},
			}
			pairs = append(pairs, newPair)
			history = append(history, newAcc)
		}
		got, ok := buildHistoryInfo(pairs)
		if ok {
			got.Lied = lieCounts[client]
			expected[client] = got
		}
	}

	return history, expected
}

func TestSaveHistoryInfo(t *testing.T) {
	cases := []struct {
		name            string
		clientVariables map[shared.ClientID][]rules.VariableFieldName
		lieCounts       map[shared.ClientID]int
		turn            uint
		reps            uint
	}{
		{
			name: "Single client simple",
			clientVariables: map[shared.ClientID][]rules.VariableFieldName{
				shared.Team1: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
				},
			},
			lieCounts: map[shared.ClientID]int{
				shared.Team1: 6,
			},
			turn: 12,
			reps: 1,
		},
		{
			name: "Multiple clients simple",
			clientVariables: map[shared.ClientID][]rules.VariableFieldName{
				shared.Team1: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
				},
				shared.Team2: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
				},
				shared.Team3: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
				},
			},
			lieCounts: map[shared.ClientID]int{
				shared.Team1: 5,
				shared.Team2: 12,
				shared.Team3: 0,
			},
			turn: 1,
			reps: 1,
		},
		{
			name: "Multiple clients mixed",
			clientVariables: map[shared.ClientID][]rules.VariableFieldName{
				shared.Team1: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
				},
				shared.Team2: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
					rules.AllocationMade,
					rules.ConstSanctionAmount,
					rules.HasIslandReportPrivateResources,
				},
				shared.Team3: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
				},
			},
			lieCounts: map[shared.ClientID]int{
				shared.Team1: 1,
				shared.Team2: 7,
				shared.Team3: 22,
			},
			turn: 13,
			reps: 1,
		},
		{
			name: "Multiple clients mixed repeat",
			clientVariables: map[shared.ClientID][]rules.VariableFieldName{
				shared.Team1: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
				},
				shared.Team2: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
					rules.ExpectedAllocation,
					rules.AllocationMade,
					rules.ConstSanctionAmount,
					rules.HasIslandReportPrivateResources,
				},
				shared.Team3: {
					rules.IslandReportedPrivateResources,
					rules.IslandActualPrivateResources,
					rules.IslandTaxContribution,
					rules.ExpectedTaxContribution,
					rules.IslandAllocation,
				},
			},
			lieCounts: map[shared.ClientID]int{
				shared.Team1: 1,
				shared.Team2: 7,
				shared.Team3: 22,
			},
			turn: 13,
			reps: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testClient := newClientInternal(id, t)
			j := testClient.clientJudge

			wholeHistory := map[uint]map[shared.ClientID]judgeHistoryInfo{}

			for i := uint(0); i < tc.reps; i++ {
				fakeHistory, expected := makeHistory(tc.clientVariables, tc.lieCounts)
				turn := tc.turn + i
				wholeHistory[turn] = expected

				j.saveHistoryInfo(&fakeHistory, &tc.lieCounts, turn)

				clientHistory := *testClient.savedHistory

				if !reflect.DeepEqual(expected, clientHistory[turn]) {
					t.Errorf("Single history failed. expected %v,\n got %v", expected, clientHistory[turn])
				}
			}

			if !reflect.DeepEqual(wholeHistory, *testClient.savedHistory) {
				t.Errorf("Whole history comparison failed. Saved history: %v", *testClient.savedHistory)
			}

		})
	}
}

func TestCallPresidentElection(t *testing.T) {
	cases := []struct {
		name               string
		monitoring         shared.MonitorResult
		turnsInPower       int
		termLength         uint
		electionRuleInPlay bool
		expectedElection   bool
	}{
		{
			name:               "no conditions",
			monitoring:         shared.MonitorResult{Performed: false, Result: false},
			turnsInPower:       1,
			termLength:         4,
			electionRuleInPlay: false,
			expectedElection:   false,
		},
		{
			name:               "term length exceeded. no rule",
			monitoring:         shared.MonitorResult{Performed: false, Result: false},
			turnsInPower:       2,
			termLength:         1,
			electionRuleInPlay: false,
			expectedElection:   false,
		},
		{
			name:               "term length exceeded. rule in play",
			monitoring:         shared.MonitorResult{Performed: false, Result: false},
			turnsInPower:       7,
			termLength:         5,
			electionRuleInPlay: true,
			expectedElection:   true,
		},
		{
			name:               "termLength=turnsInPower. rule in play",
			monitoring:         shared.MonitorResult{Performed: false, Result: false},
			turnsInPower:       5,
			termLength:         5,
			electionRuleInPlay: true,
			expectedElection:   false,
		},
		{
			name:               "monitoring done",
			monitoring:         shared.MonitorResult{Performed: true, Result: true},
			turnsInPower:       2,
			termLength:         3,
			electionRuleInPlay: true,
			expectedElection:   false,
		},
		{
			name:               "monitoring done and cheated",
			monitoring:         shared.MonitorResult{Performed: true, Result: false},
			turnsInPower:       5,
			termLength:         7,
			electionRuleInPlay: true,
			expectedElection:   true,
		},
		{
			name:               "monitoring done and cheated. term ended",
			monitoring:         shared.MonitorResult{Performed: true, Result: false},
			turnsInPower:       5,
			termLength:         4,
			electionRuleInPlay: true,
			expectedElection:   true,
		},
	}

	allTeams := []shared.ClientID{}
	for _, client := range shared.TeamIDs {
		allTeams = append(allTeams, client)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			testServer := fakeServerHandle{
				TermLengths: map[shared.Role]uint{
					shared.President: tc.termLength,
				},
				ElectionRuleInPlay: tc.electionRuleInPlay,
			}

			testClient := newClientInternal(id, t)

			testClient.Initialise(testServer)

			j := testClient.GetClientJudgePointer()

			got := j.CallPresidentElection(tc.monitoring, tc.turnsInPower, allTeams)

			if got.HoldElection != tc.expectedElection {
				t.Errorf("Expected holdElection: %v. Got holdElection: %v", tc.expectedElection, got.HoldElection)
			}
		})
	}
}

func TestGetPardonedIslands(t *testing.T) {
	cases := []struct {
		name      string
		sanctions map[int][]shared.Sanction
		pardons   map[int][]bool
		trust     []float64
	}{
		{
			name:      "empty sanctions",
			sanctions: make(map[int][]shared.Sanction),
			pardons:   make(map[int][]bool),
			trust:     make([]float64, 6),
		},
		{
			name: "no sanction",
			sanctions: map[int][]shared.Sanction{
				0: {
					{
						ClientID:     shared.Team1,
						SanctionTier: shared.NoSanction,
						TurnsLeft:    5,
					},
					{
						ClientID:     shared.Team2,
						SanctionTier: shared.NoSanction,
						TurnsLeft:    5,
					},
				},
			},
			pardons: map[int][]bool{
				0: {false, false},
			},
			trust: []float64{0.5, 0.5},
		},
		{
			name: "get pardon",
			sanctions: map[int][]shared.Sanction{
				0: {
					{
						ClientID:     shared.Team1,
						SanctionTier: shared.SanctionTier1,
						TurnsLeft:    3,
					},
				},
			},
			pardons: map[int][]bool{
				0: {true},
			},
			trust: []float64{0.7},
		},
		{
			name: "get one pardon",
			sanctions: map[int][]shared.Sanction{
				0: {
					{
						ClientID:     shared.Team1,
						SanctionTier: shared.SanctionTier1,
						TurnsLeft:    3,
					},
					{
						ClientID:     shared.Team2,
						SanctionTier: shared.SanctionTier4, // this will prevent getting a pardon
						TurnsLeft:    2,
					},
				},
			},
			pardons: map[int][]bool{
				0: {true, false},
			},
			trust: []float64{0.61, 0.99},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testServer := fakeServerHandle{}
			testClient := newClientInternal(id, t)
			testClient.Initialise(testServer)

			testClient.internalParam.agentsTrust = tc.trust

			j := testClient.GetClientJudgePointer()

			pardons := j.GetPardonedIslands(tc.sanctions)

			if reflect.DeepEqual(tc.pardons, map[int][]bool{}) {
				if !reflect.DeepEqual(pardons, map[int][]bool{}) {
					t.Errorf("GetPardonedIslands failed. expected: empty map, got %v", pardons)
				}
			}

			if !reflect.DeepEqual(pardons, tc.pardons) {
				t.Errorf("GetPardonedIslands failed. expected: %v, got %v", tc.pardons, pardons)
			}

		})
	}
}
