package polls

import (
	"time"

	"github.com/iMerica/jango/orm"
)

var QuestionMeta *orm.ModelMeta
var ChoiceMeta *orm.ModelMeta

func init() {
	QuestionMeta = orm.GlobalRegistry().Register("polls", "Question", &orm.ModelMeta{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200, orm.WithVerboseName("question text")),
			orm.DateTimeField("PubDate", orm.WithVerboseName("date published"), orm.WithDBColumn("pub_date")),
		},
		DefaultOrdering: []string{"-pub_date"},
		Options: orm.ModelOptions{
			VerboseName:        "question",
			VerboseNamePlural:  "questions",
			DefaultManagerName: "objects",
		},
	})

	ChoiceMeta = orm.GlobalRegistry().Register("polls", "Choice", &orm.ModelMeta{
		AppLabel:  "polls",
		ModelName: "Choice",
		TableName: "polls_choice",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("ChoiceText", 200, orm.WithVerboseName("choice text"), orm.WithDBColumn("choice_text")),
			orm.IntegerField("Votes", orm.WithDefault(0), orm.WithVerboseName("votes")),
			orm.ForeignKey("Question", "polls.Question", orm.WithOnDelete(orm.Cascade), orm.WithRelatedName("choices"), orm.WithDBColumn("question_id")),
		},
		DefaultOrdering: []string{},
		Options: orm.ModelOptions{
			VerboseName:        "choice",
			VerboseNamePlural:  "choices",
			DefaultManagerName: "objects",
		},
	})
}

type Question struct {
	ID           int64
	QuestionText string
	PubDate      time.Time
}

func (q *Question) TableName() string        { return "polls_question" }
func (q *Question) PKValue() interface{}     { return q.ID }
func (q *Question) SetPKValue(v interface{}) { q.ID = v.(int64) }

type Choice struct {
	ID         int64
	ChoiceText string
	Votes      int
	QuestionID int64
}

func (c *Choice) TableName() string        { return "polls_choice" }
func (c *Choice) PKValue() interface{}     { return c.ID }
func (c *Choice) SetPKValue(v interface{}) { c.ID = v.(int64) }
