package models

import "testing"

func TestStatusTransitions(t *testing.T) {
	valid := []struct{ from, to OrderStatus }{
		{StatusPendingPayment, StatusPaid},
		{StatusPendingPayment, StatusConfirmed},
		{StatusConfirmed, StatusPreparing},
		{StatusPreparing, StatusReady},
		{StatusReady, StatusOutForDelivery},
		{StatusOutForDelivery, StatusDelivered},
		{StatusDelivered, StatusRefunded},
	}
	for _, tc := range valid {
		if !tc.from.CanTransitionTo(tc.to) {
			t.Errorf("%s -> %s should be allowed", tc.from, tc.to)
		}
	}

	invalid := []struct{ from, to OrderStatus }{
		{StatusReady, StatusPaid},
		{StatusDelivered, StatusPreparing},
		{StatusCancelled, StatusConfirmed},
		{StatusPreparing, StatusDelivered}, // must pass through ready
	}
	for _, tc := range invalid {
		if tc.from.CanTransitionTo(tc.to) {
			t.Errorf("%s -> %s should NOT be allowed", tc.from, tc.to)
		}
	}
}
