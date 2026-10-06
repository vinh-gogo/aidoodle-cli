package main

import (
	"testing"
)

func TestParseCLIOptions_Flags(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantHead   bool
		wantTrends bool
		wantReview bool
		wantNext   bool
		wantErr    bool
	}{
		{
			name:       "trends and review",
			args:       []string{"--trends", "--review"},
			wantTrends: true,
			wantReview: true,
		},
		{
			name:     "headless next",
			args:     []string{"--headless", "--next"},
			wantHead: true,
			wantNext: true,
		},
		{
			name:    "next without headless fails",
			args:    []string{"--next"},
			wantErr: true,
		},
		{
			name:       "headless trends review",
			args:       []string{"--headless", "--trends", "--review"},
			wantHead:   true,
			wantTrends: true,
			wantReview: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, _, err := parseCLIOptions(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if opts.Headless != tc.wantHead {
				t.Errorf("Headless want=%v, got=%v", tc.wantHead, opts.Headless)
			}
			if opts.Trends != tc.wantTrends {
				t.Errorf("Trends want=%v, got=%v", tc.wantTrends, opts.Trends)
			}
			if opts.Review != tc.wantReview {
				t.Errorf("Review want=%v, got=%v", tc.wantReview, opts.Review)
			}
			if opts.Next != tc.wantNext {
				t.Errorf("Next want=%v, got=%v", tc.wantNext, opts.Next)
			}
		})
	}
}
