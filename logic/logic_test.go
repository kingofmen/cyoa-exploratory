package logic

import (
	"testing"

	"google.golang.org/protobuf/proto"

	lpb "github.com/kingofmen/cyoa-exploratory/logic/proto"
)

func TestBasics(t *testing.T) {
	defaults := NewTestLookup().
		WithInt("one", 1).
		WithInt("en", 1).
		WithInt("two", 2).
		WithStr("string1", "yohoho").
		WithStr("string2", "yohoho").
		WithStr("string3", "bwahaha").
		WithStrArr("strarr1", []string{"yohoho", "bwahaha"})

	cases := []struct {
		desc   string
		pred   *lpb.Predicate
		lookup Lookup
		want   bool
	}{
		{
			desc:   "Nil predicate is true",
			lookup: defaults,
			want:   true,
		},
		{
			desc:   "Unconditional predicate is true",
			pred:   &lpb.Predicate{},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Greater than (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("two"),
						Operation: lpb.Compare_CMP_GT.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Greater than (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_GT.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Less than (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_LT.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Less than (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("two"),
						Operation: lpb.Compare_CMP_LT.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Equal (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_EQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Equal (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("en"),
						Operation: lpb.Compare_CMP_EQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Greater than or equal (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("two"),
						Operation: lpb.Compare_CMP_GTE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Greater than or equal (true, greater)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_GTE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Greater than or equal (true, equal)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("en"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_GTE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Less than or equal (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_LTE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Less than or equal (true, less than)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("two"),
						Operation: lpb.Compare_CMP_LTE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Less than or equal (true, equal)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("en"),
						Operation: lpb.Compare_CMP_LTE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Not equal (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("one"),
						Operation: lpb.Compare_CMP_NEQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Not equal (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("one"),
						KeyTwo:    proto.String("en"),
						Operation: lpb.Compare_CMP_NEQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "String equals (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("string1"),
						KeyTwo:    proto.String("string3"),
						Operation: lpb.Compare_CMP_STREQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "String equals (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("string1"),
						KeyTwo:    proto.String("string2"),
						Operation: lpb.Compare_CMP_STREQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Integer literals",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("1"),
						KeyTwo:    proto.String("2"),
						Operation: lpb.Compare_CMP_GT.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Integer literal mixed with variable",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("two"),
						KeyTwo:    proto.String("1"),
						Operation: lpb.Compare_CMP_GT.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "String literals",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("'literal"),
						KeyTwo:    proto.String("'another literal"),
						Operation: lpb.Compare_CMP_STREQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "String literal mixed with lookup",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("'yohoho"),
						KeyTwo:    proto.String("string1"),
						Operation: lpb.Compare_CMP_STREQ.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "String in array (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("'banana"),
						KeyTwo:    proto.String("strarr1"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "String in array (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("string1"),
						KeyTwo:    proto.String("strarr1"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "String in array literal (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("'carrot"),
						KeyTwo:    proto.String("['apple, 'banana, string1]"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "String in array literal (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("'yohoho"),
						KeyTwo:    proto.String("['apple, 'banana, string1]"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			got, err := Eval(cc.pred, cc.lookup)
			if err != nil {
				t.Errorf("%s: Eval() => %v, want nil", cc.desc, err)
			}
			if got != cc.want {
				t.Errorf("%s: Eval() => %v, want %v", cc.desc, got, cc.want)
			}
		})
	}
}

func TestCombinations(t *testing.T) {
	defaults := NewTestLookup().
		WithInt("one", 1).
		WithInt("en", 1).
		WithInt("two", 2)

	cases := []struct {
		desc   string
		pred   *lpb.Predicate
		lookup Lookup
		want   bool
	}{
		{
			desc: "All (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comb{
					Comb: &lpb.Combine{
						Operands: []*lpb.Predicate{
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_LT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("two"),
										KeyTwo:    proto.String("one"),
										Operation: lpb.Compare_CMP_GT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("en"),
										Operation: lpb.Compare_CMP_EQ.Enum(),
									},
								},
							},
						},
						Operation: lpb.Combine_IF_ALL.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "All (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comb{
					Comb: &lpb.Combine{
						Operands: []*lpb.Predicate{
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_LT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("two"),
										KeyTwo:    proto.String("one"),
										Operation: lpb.Compare_CMP_GT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_EQ.Enum(),
									},
								},
							},
						},
						Operation: lpb.Combine_IF_ALL.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "Any (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comb{
					Comb: &lpb.Combine{
						Operands: []*lpb.Predicate{
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_GT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("two"),
										KeyTwo:    proto.String("one"),
										Operation: lpb.Compare_CMP_LT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_NEQ.Enum(),
									},
								},
							},
						},
						Operation: lpb.Combine_IF_ANY.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
		{
			desc: "Any (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comb{
					Comb: &lpb.Combine{
						Operands: []*lpb.Predicate{
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_GTE.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("two"),
										KeyTwo:    proto.String("one"),
										Operation: lpb.Compare_CMP_LTE.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_EQ.Enum(),
									},
								},
							},
						},
						Operation: lpb.Combine_IF_ANY.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "None (false)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comb{
					Comb: &lpb.Combine{
						Operands: []*lpb.Predicate{
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_GT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("two"),
										KeyTwo:    proto.String("one"),
										Operation: lpb.Compare_CMP_LT.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_NEQ.Enum(),
									},
								},
							},
						},
						Operation: lpb.Combine_IF_NONE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   false,
		},
		{
			desc: "None (true)",
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comb{
					Comb: &lpb.Combine{
						Operands: []*lpb.Predicate{
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_GTE.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("two"),
										KeyTwo:    proto.String("one"),
										Operation: lpb.Compare_CMP_LTE.Enum(),
									},
								},
							},
							&lpb.Predicate{
								Test: &lpb.Predicate_Comp{
									Comp: &lpb.Compare{
										KeyOne:    proto.String("one"),
										KeyTwo:    proto.String("two"),
										Operation: lpb.Compare_CMP_EQ.Enum(),
									},
								},
							},
						},
						Operation: lpb.Combine_IF_NONE.Enum(),
					},
				},
			},
			lookup: defaults,
			want:   true,
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			got, err := Eval(cc.pred, cc.lookup)
			if err != nil {
				t.Errorf("%s: Eval() => %v, want nil", cc.desc, err)
			}
			if got != cc.want {
				t.Errorf("%s: Eval() => %v, want %v", cc.desc, got, cc.want)
			}
		})
	}
}

func TestScopes(t *testing.T) {
	cases := []struct {
		desc   string
		base   *TestLookup
		scopes map[string]*TestLookup
		pred   *lpb.Predicate
		want   bool
	}{
		{
			desc: "Compare base and scope",
			base: NewTestLookup().WithStr("something", "abc"),
			scopes: map[string]*TestLookup{
				"scope": NewTestLookup().WithStr("another", "abc"),
			},
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("something"),
						KeyTwo:    proto.String("scope.another"),
						Operation: lpb.Compare_CMP_STREQ.Enum(),
					},
				},
			},
			want: true,
		},
		{
			desc: "Base string in scope array",
			base: NewTestLookup().WithStr("something", "abc"),
			scopes: map[string]*TestLookup{
				"scope": NewTestLookup().WithStrArr("another", []string{"abc", "def"}),
			},
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("something"),
						KeyTwo:    proto.String("scope.another"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			want: true,
		},
		{
			desc: "String literal in scope array",
			base: NewTestLookup(),
			scopes: map[string]*TestLookup{
				"scope": NewTestLookup().WithStrArr("another", []string{"abc", "def"}),
			},
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("'def"),
						KeyTwo:    proto.String("scope.another"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			want: true,
		},
		{
			desc: "String from one scope in array from another",
			base: NewTestLookup(),
			scopes: map[string]*TestLookup{
				"tele":  NewTestLookup().WithStr("foo", "abc"),
				"scope": NewTestLookup().WithStrArr("another", []string{"abc", "def"}),
			},
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("tele.foo"),
						KeyTwo:    proto.String("scope.another"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			want: true,
		},
		{
			desc: "String from scope in array literal",
			base: NewTestLookup(),
			scopes: map[string]*TestLookup{
				"tele": NewTestLookup().WithStr("foo", "abc"),
			},
			pred: &lpb.Predicate{
				Test: &lpb.Predicate_Comp{
					Comp: &lpb.Compare{
						KeyOne:    proto.String("tele.foo"),
						KeyTwo:    proto.String("['abc, 'def]"),
						Operation: lpb.Compare_CMP_STRIN.Enum(),
					},
				},
			},
			want: true,
		},
	}

	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			for key, scope := range cc.scopes {
				cc.base.SetScope(key, scope)
			}
			got, err := Eval(cc.pred, cc.base)
			if err != nil {
				t.Fatalf("%s: Eval() => %v, want nil", cc.desc, err)
			}
			if got != cc.want {
				t.Errorf("%s: Eval() => %v, want %v", cc.desc, got, cc.want)
			}
		})
	}
}
