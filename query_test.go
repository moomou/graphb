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

	t.Run("field filter", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeDgraphQuery).
			SetFields(
				NewFuncField("s").SetArguments(
					ArgumentFuncType(
						"eq",
						ArgumentString("name@en", "RidleyScott"),
					),
				).SetFields(
					NewField("director.film@filter").SetArguments(
						ArgumentFuncType(
							"le",
							ArgumentString("initial_release_date", "2000"),
						),
					).SetFields(
						NewField("initial_release_date"),
					),
				),
			)

		assert.Equal(t, `{s(func:eq(name@en,"RidleyScott")){director.film@filter(le(initial_release_date,"2000")){initial_release_date}}}`, buildString(q.stringChan()))
	})

	t.Run("boolean", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeDgraphQuery).
			SetFields(
				NewFuncField("me").SetArguments(
					ArgumentFuncType(
						"eq",
						ArgumentString("name@en", "StevenSpielberg"),
					),
				).SetFields(
					NewField("name@en@filter").SetArguments(
						ArgumentFuncType(
							"has",
							ArgumentString("director.film", ""),
						),
					),
					NewField("director.film@filter").SetArgumentsWithBool(
						[]string{"OR"},
						ArgumentFuncType(
							"allofterms",
							ArgumentString("name@en", "jonesindiana"),
						),
						ArgumentFuncType(
							"allofterms",
							ArgumentString("name@en", "jurassicpark"),
						),
					).SetFields(
						NewField("uid"),
					),
				),
			)

		assert.Equal(t, `{me(func:eq(name@en,"StevenSpielberg")){name@en@filter(has(director.film)),director.film@filter(allofterms(name@en,"jonesindiana") OR allofterms(name@en,"jurassicpark")){uid}}}`, buildString(q.stringChan()))
	})

	t.Run("query filter wt boolean", func(t *testing.T) {
		t.Parallel()

		q := NewQuery(TypeDgraphQuery).
			SetFields(
				NewFuncField("query").SetArguments(
					ArgumentFuncType(
						"eq",
						ArgumentString("word@en", "dog"),
					),
				).Filter(
					[]string{"AND"},
					ArgumentFuncType(
						"has",
						ArgumentString("name", ""),
					),
					ArgumentFuncType(
						"eq",
						ArgumentString("age", "7"),
					),
				).SetFields(
					NewField("name"),
					NewField("director.film").SetFields(
						NewField("uid"),
					),
				),
			)

		assert.Equal(t, `{query(func:eq(word@en,"dog"))@filter(has(name) AND eq(age,"7")){name,director.film{uid}}}`, buildString(q.stringChan()))
	})

}
