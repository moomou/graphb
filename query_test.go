package graphb

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestQuery_checkName(t *testing.T) {
	q := Query{Name: "1"}
	err := q.checkName()
	assert.IsType(t, InvalidNameErr{}, errors.Cause(err))
	assert.Equal(t, "'1' is an invalid operation name in GraphQL. A valid name matches /[_A-Za-z][_0-9A-Za-z]*/, see: http://facebook.github.io/graphql/October2016/#sec-Names", err.Error())
}

func TestQuery_check(t *testing.T) {
	q := Query{Name: "1", Type: TypeQuery}
	err := q.check()
	assert.IsType(t, InvalidNameErr{}, errors.Cause(err))
	assert.Equal(t, "'1' is an invalid operation name in GraphQL. A valid name matches /[_A-Za-z][_0-9A-Za-z]*/, see: http://facebook.github.io/graphql/October2016/#sec-Names", err.Error())
}

func TestQuery_GetField(t *testing.T) {
	q := MakeQuery(TypeQuery).SetFields(MakeField("f1"))
	f := q.GetField("f1")
	assert.Equal(t, "f1", f.Name)

	f = q.GetField("f2")
	assert.Nil(t, f)
}

func TestQuery_JSON(t *testing.T) {
	t.Parallel()

	t.Run("Arguments can be nested structures", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeMutation).
			SetFields(
				NewField("createQuestion").
					SetArguments(
						ArgumentCustomType(
							"input",
							ArgumentString("title", "what"),
							ArgumentString("content", "what"),
							ArgumentStringSlice("tagIds"),
						),
					).
					SetFields(
						NewField("question", OfFields("id")),
					),
			)

		c := q.stringChan()

		var strs []string
		for str := range c {
			strs = append(strs, str)
		}

		assert.Equal(t, `mutation{createQuestion(input:{title:"what",content:"what",tagIds:[]}){question{id}}}`, strings.Join(strs, ""))
	})

}

func TestDgraphQuery(t *testing.T) {
	buildString := func(strChan <-chan string) string {
		var strs []string
		for str := range strChan {
			strs = append(strs, str)
		}
		return strings.Join(strs, "")
	}

	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeDgraphQuery).
			SetName("").
			SetFields(
				NewFuncField("all").SetArguments(
					ArgumentFuncType(
						"anyofterms",
						ArgumentString("name", "blob"),
					),
				).SetFields(
					NewField("question"),
				),
			)

		assert.Equal(t, `{all(func:anyofterms(name,"blob")){question}}`, buildString(q.stringChan()))
	})

	t.Run("simple2", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeDgraphQuery).
			SetName("").
			SetFields(
				NewFuncField("bd").SetArguments(
					ArgumentFuncType(
						"eq",
						ArgumentString("name@en", "Blade Runner"),
					),
				).SetFields(
					NewField("uid"),
					NewField("name@en"),
					NewField("initial_release_date"),
					NewField("netflix_id"),
				),
			)

		assert.Equal(t, `{bd(func:eq(name@en,"Blade Runner")){uid,name@en,initial_release_date,netflix_id}}`, buildString(q.stringChan()))
	})

	t.Run("nested", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeDgraphQuery).
			SetName("").
			SetFields(
				NewFuncField("mf").SetArguments(
					ArgumentFuncType(
						"eq",
						ArgumentString("name", "Michael"),
					),
				).SetFields(
					NewField("name"),
					NewField("age"),
					NewField("friend").SetFields(
						NewField("name@."),
					),
				),
			)

		assert.Equal(t, `{mf(func:eq(name,"Michael")){name,age,friend{name@.}}}`, buildString(q.stringChan()))
	})

}
