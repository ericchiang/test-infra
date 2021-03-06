/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package approvers

import (
	"testing"

	"reflect"

	"k8s.io/kubernetes/pkg/util/sets"
)

func TestUnapprovedFiles(t *testing.T) {
	rootApprovers := sets.NewString("Alice", "Bob")
	aApprovers := sets.NewString("Art", "Anne")
	bApprovers := sets.NewString("Bill", "Ben", "Barbara")
	cApprovers := sets.NewString("Chris", "Carol")
	dApprovers := sets.NewString("David", "Dan", "Debbie")
	eApprovers := sets.NewString("Eve", "Erin")
	edcApprovers := eApprovers.Union(dApprovers).Union(cApprovers)
	FakeRepoMap := map[string]sets.String{
		"":        rootApprovers,
		"a":       aApprovers,
		"b":       bApprovers,
		"c":       cApprovers,
		"a/d":     dApprovers,
		"a/combo": edcApprovers,
	}
	tests := []struct {
		testName           string
		filenames          []string
		currentlyApproved  sets.String
		expectedUnapproved sets.String
	}{
		{
			testName:           "Empty PR",
			filenames:          []string{},
			currentlyApproved:  sets.NewString(),
			expectedUnapproved: sets.NewString(),
		},
		{
			testName:           "Single Root File PR Approved",
			filenames:          []string{"kubernetes.go"},
			currentlyApproved:  sets.NewString(rootApprovers.List()[0]),
			expectedUnapproved: sets.NewString(),
		},
		{
			testName:           "Single Root File PR No One Approved",
			filenames:          []string{"kubernetes.go"},
			currentlyApproved:  sets.NewString(),
			expectedUnapproved: sets.NewString(""),
		},
		{
			testName:           "B Only UnApproved",
			currentlyApproved:  bApprovers,
			expectedUnapproved: sets.NewString(),
		},
		{
			testName:           "B Files Approved at Root",
			filenames:          []string{"b/test.go", "b/test_1.go"},
			currentlyApproved:  rootApprovers,
			expectedUnapproved: sets.NewString(),
		},
		{
			testName:           "B Only UnApproved",
			filenames:          []string{"b/test_1.go", "b/test.go"},
			currentlyApproved:  sets.NewString(),
			expectedUnapproved: sets.NewString("b"),
		},
		{
			testName:           "Combo and Other; Neither Approved",
			filenames:          []string{"a/combo/test.go", "a/d/test.go"},
			currentlyApproved:  sets.NewString(),
			expectedUnapproved: sets.NewString("a/combo", "a/d"),
		},
		{
			testName:           "Combo and Other; Combo Approved",
			filenames:          []string{"a/combo/test.go", "a/d/test.go"},
			currentlyApproved:  edcApprovers.Difference(dApprovers),
			expectedUnapproved: sets.NewString("a/d"),
		},
		{
			testName:           "Combo and Other; Both Approved",
			filenames:          []string{"a/combo/test.go", "a/d/test.go"},
			currentlyApproved:  edcApprovers.Intersection(dApprovers),
			expectedUnapproved: sets.NewString(),
		},
	}

	for _, test := range tests {
		testApprovers := NewApprovers(Owners{filenames: test.filenames, repo: createFakeRepo(FakeRepoMap), seed: TEST_SEED})
		for approver := range test.currentlyApproved {
			testApprovers.AddApprover(approver, "REFERENCE")
		}
		calculated := testApprovers.UnapprovedFiles()
		if !test.expectedUnapproved.Equal(calculated) {
			t.Errorf("Failed for test %v.  Expected unapproved files: %v. Found %v", test.testName, test.expectedUnapproved, calculated)
		}
	}
}

func TestGetFiles(t *testing.T) {
	rootApprovers := sets.NewString("Alice", "Bob")
	aApprovers := sets.NewString("Art", "Anne")
	bApprovers := sets.NewString("Bill", "Ben", "Barbara")
	cApprovers := sets.NewString("Chris", "Carol")
	dApprovers := sets.NewString("David", "Dan", "Debbie")
	eApprovers := sets.NewString("Eve", "Erin")
	edcApprovers := eApprovers.Union(dApprovers).Union(cApprovers)
	FakeRepoMap := map[string]sets.String{
		"":        rootApprovers,
		"a":       aApprovers,
		"b":       bApprovers,
		"c":       cApprovers,
		"a/d":     dApprovers,
		"a/combo": edcApprovers,
	}
	tests := []struct {
		testName          string
		filenames         []string
		currentlyApproved sets.String
		expectedFiles     []File
	}{
		{
			testName:          "Empty PR",
			filenames:         []string{},
			currentlyApproved: sets.NewString(),
			expectedFiles:     []File{},
		},
		{
			testName:          "Single Root File PR Approved",
			filenames:         []string{"kubernetes.go"},
			currentlyApproved: sets.NewString(rootApprovers.List()[0]),
			expectedFiles:     []File{ApprovedFile{"", sets.NewString(rootApprovers.List()[0]), "org", "project"}},
		},
		{
			testName:          "Single File PR in B No One Approved",
			filenames:         []string{"b/test.go"},
			currentlyApproved: sets.NewString(),
			expectedFiles:     []File{UnapprovedFile{"b", "org", "project"}},
		},
		{
			testName:          "Single File PR in B Fully Approved",
			filenames:         []string{"b/test.go"},
			currentlyApproved: bApprovers,
			expectedFiles:     []File{ApprovedFile{"b", bApprovers, "org", "project"}},
		},
		{
			testName:          "Single Root File PR No One Approved",
			filenames:         []string{"kubernetes.go"},
			currentlyApproved: sets.NewString(),
			expectedFiles:     []File{UnapprovedFile{"", "org", "project"}},
		},
		{
			testName:          "Combo and Other; Neither Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			currentlyApproved: sets.NewString(),
			expectedFiles: []File{
				UnapprovedFile{"a/combo", "org", "project"},
				UnapprovedFile{"a/d", "org", "project"},
			},
		},
		{
			testName:          "Combo and Other; Combo Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			currentlyApproved: eApprovers,
			expectedFiles: []File{
				ApprovedFile{"a/combo", eApprovers, "org", "project"},
				UnapprovedFile{"a/d", "org", "project"},
			},
		},
		{
			testName:          "Combo and Other; Both Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			currentlyApproved: edcApprovers.Intersection(dApprovers),
			expectedFiles: []File{
				ApprovedFile{"a/combo", edcApprovers.Intersection(dApprovers), "org", "project"},
				ApprovedFile{"a/d", edcApprovers.Intersection(dApprovers), "org", "project"},
			},
		},
		{
			testName:          "Combo, C, D; Combo and C Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go", "c/test"},
			currentlyApproved: cApprovers,
			expectedFiles: []File{
				ApprovedFile{"a/combo", cApprovers, "org", "project"},
				UnapprovedFile{"a/d", "org", "project"},
				ApprovedFile{"c", cApprovers, "org", "project"},
			},
		},
		{
			testName:          "Files Approved Multiple times",
			filenames:         []string{"a/test.go", "a/d/test.go", "b/test"},
			currentlyApproved: rootApprovers.Union(aApprovers).Union(bApprovers),
			expectedFiles: []File{
				ApprovedFile{"a", rootApprovers.Union(aApprovers), "org", "project"},
				ApprovedFile{"b", rootApprovers.Union(bApprovers), "org", "project"},
			},
		},
	}

	for _, test := range tests {
		testApprovers := NewApprovers(Owners{filenames: test.filenames, repo: createFakeRepo(FakeRepoMap), seed: TEST_SEED})
		for approver := range test.currentlyApproved {
			testApprovers.AddApprover(approver, "REFERENCE")
		}
		calculated := testApprovers.GetFiles("org", "project")
		if !reflect.DeepEqual(test.expectedFiles, calculated) {
			t.Errorf("Failed for test %v.  Expected files: %v. Found %v", test.testName, test.expectedFiles, calculated)
		}
	}
}

func TestGetCCs(t *testing.T) {
	rootApprovers := sets.NewString("Alice", "Bob")
	aApprovers := sets.NewString("Art", "Anne")
	bApprovers := sets.NewString("Bill", "Ben", "Barbara")
	cApprovers := sets.NewString("Chris", "Carol")
	dApprovers := sets.NewString("David", "Dan", "Debbie")
	eApprovers := sets.NewString("Eve", "Erin")
	edcApprovers := eApprovers.Union(dApprovers).Union(cApprovers)
	FakeRepoMap := map[string]sets.String{
		"":        rootApprovers,
		"a":       aApprovers,
		"b":       bApprovers,
		"c":       cApprovers,
		"a/d":     dApprovers,
		"a/combo": edcApprovers,
	}
	tests := []struct {
		testName          string
		filenames         []string
		currentlyApproved sets.String
		// testSeed affects who is chosen for CC
		testSeed  int64
		assignees []string
		// order matters for CCs
		expectedCCs []string
	}{
		{
			testName:          "Empty PR",
			filenames:         []string{},
			currentlyApproved: sets.NewString(),
			testSeed:          0,
			expectedCCs:       []string{},
		},
		{
			testName:          "Single Root FFile PR Approved",
			filenames:         []string{"kubernetes.go"},
			currentlyApproved: sets.NewString(rootApprovers.List()[0]),
			testSeed:          13,
			expectedCCs:       []string{},
		},
		{
			testName:          "Single Root File PR Unapproved Seed = 13",
			filenames:         []string{"kubernetes.go"},
			currentlyApproved: sets.NewString(),
			testSeed:          13,
			expectedCCs:       []string{"Alice"},
		},
		{
			testName:          "Single Root File PR No One Seed = 10",
			filenames:         []string{"kubernetes.go"},
			testSeed:          10,
			currentlyApproved: sets.NewString(),
			expectedCCs:       []string{"Bob"},
		},
		{
			testName:          "Combo and Other; Neither Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			expectedCCs:       []string{"Dan"},
		},
		{
			testName:          "Combo and Other; Combo Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			testSeed:          0,
			currentlyApproved: eApprovers,
			expectedCCs:       []string{"Dan"},
		},
		{
			testName:          "Combo and Other; Both Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			testSeed:          0,
			currentlyApproved: dApprovers, // dApprovers can approve combo and d directory
			expectedCCs:       []string{},
		},
		{
			testName:          "Combo, C, D; None Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			// chris can approve c and combo, debbie can approve d
			expectedCCs: []string{"Chris", "Debbie"},
		},
		{
			testName:          "A, B, C; Nothing Approved",
			filenames:         []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			// Need an approver from each of the three owners files
			expectedCCs: []string{"Anne", "Bill", "Carol"},
		},
		{
			testName:  "A, B, C; Partially approved by non-suggested approvers",
			filenames: []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:  0,
			// Approvers are valid approvers, but not the one we would suggest
			currentlyApproved: sets.NewString("Art", "Ben"),
			// We don't suggest approvers for a and b, only for unapproved c.
			expectedCCs: []string{"Carol"},
		},
		{
			testName:  "A, B, C; Nothing approved, but assignees can approve",
			filenames: []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:  0,
			// Approvers are valid approvers, but not the one we would suggest
			currentlyApproved: sets.NewString(),
			assignees:         []string{"Art", "Ben"},
			// We suggest assigned people rather than "suggested" people
			// Suggested would be "Anne", "Bill", "Carol" if no one was assigned.
			expectedCCs: []string{"Art", "Ben", "Carol"},
		},
		{
			testName:          "A, B, C; Nothing approved, but SOME assignees can approve",
			filenames:         []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			// Assignees are a mix of potential approvers and random people
			assignees: []string{"Art", "Ben", "John", "Jack"},
			// We suggest assigned people rather than "suggested" people
			expectedCCs: []string{"Art", "Ben", "Carol"},
		},
		{
			testName:          "Assignee is top OWNER, No one has approved",
			filenames:         []string{"a/test.go"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			// Assignee is a root approver
			assignees:   []string{"Alice"},
			expectedCCs: []string{"Alice"},
		},
	}

	for _, test := range tests {
		testApprovers := NewApprovers(Owners{filenames: test.filenames, repo: createFakeRepo(FakeRepoMap), seed: test.testSeed})
		for approver := range test.currentlyApproved {
			testApprovers.AddApprover(approver, "REFERENCE")
		}
		testApprovers.AddAssignees(test.assignees...)
		calculated := testApprovers.GetCCs()
		if !reflect.DeepEqual(test.expectedCCs, calculated) {
			t.Errorf("Failed for test %v.  Expected CCs: %v. Found %v", test.testName, test.expectedCCs, calculated)
		}
	}
}

func TestIsApproved(t *testing.T) {
	rootApprovers := sets.NewString("Alice", "Bob")
	aApprovers := sets.NewString("Art", "Anne")
	bApprovers := sets.NewString("Bill", "Ben", "Barbara")
	cApprovers := sets.NewString("Chris", "Carol")
	dApprovers := sets.NewString("David", "Dan", "Debbie")
	eApprovers := sets.NewString("Eve", "Erin")
	edcApprovers := eApprovers.Union(dApprovers).Union(cApprovers)
	FakeRepoMap := map[string]sets.String{
		"":        rootApprovers,
		"a":       aApprovers,
		"b":       bApprovers,
		"c":       cApprovers,
		"a/d":     dApprovers,
		"a/combo": edcApprovers,
	}
	tests := []struct {
		testName          string
		filenames         []string
		currentlyApproved sets.String
		testSeed          int64
		isApproved        bool
	}{
		{
			testName:          "Empty PR",
			filenames:         []string{},
			currentlyApproved: sets.NewString(),
			testSeed:          0,
			isApproved:        true,
		},
		{
			testName:          "Single Root File PR Approved",
			filenames:         []string{"kubernetes.go"},
			currentlyApproved: sets.NewString(rootApprovers.List()[0]),
			testSeed:          3,
			isApproved:        true,
		},
		{
			testName:          "Single Root File PR No One Approved",
			filenames:         []string{"kubernetes.go"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			isApproved:        false,
		},
		{
			testName:          "Combo and Other; Neither Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			isApproved:        false,
		},
		{
			testName:          "Combo and Other; Both Approved",
			filenames:         []string{"a/combo/test.go", "a/d/test.go"},
			testSeed:          0,
			currentlyApproved: edcApprovers.Intersection(dApprovers),
			isApproved:        true,
		},
		{
			testName:          "A, B, C; Nothing Approved",
			filenames:         []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: sets.NewString(),
			isApproved:        false,
		},
		{
			testName:          "A, B, C; Partially Approved",
			filenames:         []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: aApprovers.Union(bApprovers),
			isApproved:        false,
		},
		{
			testName:          "A, B, C; Approved At the Root",
			filenames:         []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: rootApprovers,
			isApproved:        true,
		},
		{
			testName:          "A, B, C; Approved At the Leaves",
			filenames:         []string{"a/test.go", "b/test.go", "c/test"},
			testSeed:          0,
			currentlyApproved: sets.NewString("Anne", "Ben", "Carol"),
			isApproved:        true,
		},
	}

	for _, test := range tests {
		testApprovers := NewApprovers(Owners{filenames: test.filenames, repo: createFakeRepo(FakeRepoMap), seed: test.testSeed})
		for approver := range test.currentlyApproved {
			testApprovers.AddApprover(approver, "REFERENCE")
		}
		calculated := testApprovers.IsApproved()
		if test.isApproved != calculated {
			t.Errorf("Failed for test %v.  Expected Approval Status: %v. Found %v", test.testName, test.isApproved, calculated)
		}
	}
}

func TestGetFilesApprovers(t *testing.T) {
	tests := []struct {
		testName       string
		filenames      []string
		approvers      []string
		owners         map[string]sets.String
		expectedStatus map[string]sets.String
	}{
		{
			testName:       "Empty PR",
			filenames:      []string{},
			approvers:      []string{},
			owners:         map[string]sets.String{},
			expectedStatus: map[string]sets.String{},
		},
		{
			testName:       "No approvers",
			filenames:      []string{"a/a", "c"},
			approvers:      []string{},
			owners:         map[string]sets.String{"": sets.NewString("RootOwner")},
			expectedStatus: map[string]sets.String{"": sets.String{}},
		},
		{
			testName: "Approvers approves some",
			filenames: []string{
				"a/a",
				"c/c",
			},
			approvers: []string{"CApprover"},
			owners: map[string]sets.String{
				"a": sets.NewString("AApprover"),
				"c": sets.NewString("CApprover"),
			},
			expectedStatus: map[string]sets.String{
				"a": sets.String{},
				"c": sets.NewString("CApprover"),
			},
		},
		{
			testName: "Multiple approvers",
			filenames: []string{
				"a/a",
				"c/c",
			},
			approvers: []string{"RootApprover", "CApprover"},
			owners: map[string]sets.String{
				"":  sets.NewString("RootApprover"),
				"a": sets.NewString("AApprover"),
				"c": sets.NewString("CApprover"),
			},
			expectedStatus: map[string]sets.String{
				"a": sets.NewString("RootApprover"),
				"c": sets.NewString("RootApprover", "CApprover"),
			},
		},
		{
			testName:       "Case-insensitive approvers",
			filenames:      []string{"file"},
			approvers:      []string{"RootApprover"},
			owners:         map[string]sets.String{"": sets.NewString("rOOtaPProver")},
			expectedStatus: map[string]sets.String{"": sets.NewString("RootApprover")},
		},
	}

	for _, test := range tests {
		testApprovers := NewApprovers(Owners{filenames: test.filenames, repo: createFakeRepo(test.owners)})
		for _, approver := range test.approvers {
			testApprovers.AddApprover(approver, "REFERENCE")
		}
		calculated := testApprovers.GetFilesApprovers()
		if !reflect.DeepEqual(test.expectedStatus, calculated) {
			t.Errorf("Failed for test %v.  Expected approval status: %v. Found %v", test.testName, test.expectedStatus, calculated)
		}
	}
}

func TestGetMessage(t *testing.T) {
	ap := NewApprovers(
		Owners{
			filenames: []string{"a/a.go", "b/b.go"},
			repo: createFakeRepo(map[string]sets.String{
				"a": sets.NewString("Alice"),
				"b": sets.NewString("Bill"),
			}),
		},
	)
	ap.AddApprover("Bill", "REFERENCE")

	want := `[APPROVALNOTIFIER] This PR is **NOT APPROVED**

This pull-request has been approved by: *<a href="REFERENCE" title="Approved">Bill</a>*
We suggest the following additional approver: **Alice**

Assign the PR to them by writing ` + "`/assign @Alice`" + ` in a comment when ready.

<details open>
Needs approval from an approver in each of these OWNERS Files:

- **[a/OWNERS](https://github.com/org/project/blob/master/a/OWNERS)**
- ~~[b/OWNERS](https://github.com/org/project/blob/master/b/OWNERS)~~ [Bill]

You can indicate your approval by writing ` + "`/approve`" + ` in a comment
You can cancel your approval by writing ` + "`/approve cancel`" + ` in a comment
</details>
<!-- META={"approvers":["Alice"]} -->`
	if got := GetMessage(ap, "org", "project"); got == nil {
		t.Error("GetMessage() failed")
	} else if *got != want {
		t.Errorf("GetMessage() = %+v, want = %+v", *got, want)
	}
}

func TestGetMessageAllApproved(t *testing.T) {
	ap := NewApprovers(
		Owners{
			filenames: []string{"a/a.go", "b/b.go"},
			repo: createFakeRepo(map[string]sets.String{
				"a": sets.NewString("Alice"),
				"b": sets.NewString("Bill"),
			}),
		},
	)
	ap.AddApprover("Alice", "REFERENCE")
	ap.AddLGTMer("Bill", "REFERENCE")

	want := `[APPROVALNOTIFIER] This PR is **APPROVED**

This pull-request has been approved by: *<a href="REFERENCE" title="Approved">Alice</a>*, *<a href="REFERENCE" title="LGTM">Bill</a>*

<details >
Needs approval from an approver in each of these OWNERS Files:

- ~~[a/OWNERS](https://github.com/org/project/blob/master/a/OWNERS)~~ [Alice]
- ~~[b/OWNERS](https://github.com/org/project/blob/master/b/OWNERS)~~ [Bill]

You can indicate your approval by writing ` + "`/approve`" + ` in a comment
You can cancel your approval by writing ` + "`/approve cancel`" + ` in a comment
</details>
<!-- META={"approvers":[]} -->`
	if got := GetMessage(ap, "org", "project"); got == nil {
		t.Error("GetMessage() failed")
	} else if *got != want {
		t.Errorf("GetMessage() = %+v, want = %+v", *got, want)
	}
}

func TestGetMessageNoneApproved(t *testing.T) {
	ap := NewApprovers(
		Owners{
			filenames: []string{"a/a.go", "b/b.go"},
			repo: createFakeRepo(map[string]sets.String{
				"a": sets.NewString("Alice"),
				"b": sets.NewString("Bill"),
			}),
		},
	)

	want := `[APPROVALNOTIFIER] This PR is **NOT APPROVED**

This pull-request has been approved by: 
We suggest the following additional approvers: **Alice**, **Bill**

Assign the PR to them by writing ` + "`/assign @Alice @Bill`" + ` in a comment when ready.

<details open>
Needs approval from an approver in each of these OWNERS Files:

- **[a/OWNERS](https://github.com/org/project/blob/master/a/OWNERS)**
- **[b/OWNERS](https://github.com/org/project/blob/master/b/OWNERS)**

You can indicate your approval by writing ` + "`/approve`" + ` in a comment
You can cancel your approval by writing ` + "`/approve cancel`" + ` in a comment
</details>
<!-- META={"approvers":["Alice","Bill"]} -->`
	if got := GetMessage(ap, "org", "project"); got == nil {
		t.Error("GetMessage() failed")
	} else if *got != want {
		t.Errorf("GetMessage() = %+v, want = %+v", *got, want)
	}
}
