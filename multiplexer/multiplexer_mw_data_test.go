package multiplexer

import "go.mongodb.org/mongo-driver/bson/primitive"

// TestUser for middleware test
type TestUser struct {
	ID    primitive.ObjectID `json:"-" bson:"_id" _id_:"user"`
	Name  string             `json:"name" _hd_:"c"`
	Email string             `json:"email" _hd_:"c"`
	//Age   int64              `json:"age" _hd_:"c"`
}

var DummyUserData = TestUser{Name: "Dummy UserEmbed", Email: "dummy@user.com"}

const DummyUserDataJSON = `{"name": "Dummy UserEmbed","email": "dummy@user.com"}`

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TaskDetails struct {
	Date string `json:"date" _id_:"task-details" _hd_:"c"`
}

type Task struct {
	Name    string      `json:"name" _id_:"task" _hd_:"c"`
	Details TaskDetails `json:"details" _hd_:"c"`
}

type UserEmbed struct {
	Tasks Task `json:"tasks" _id_:"user-embed" _hd_:"c"`
}

var DummyUserEmbed = UserEmbed{
	Tasks: Task{
		Name:    "test task",
		Details: TaskDetails{Date: "ISO_DUMMY_DATE"},
	},
}

const dummyEmbedDataJSON = `{
  "tasks": {
    "name": "test task",
    "details": {
      "date": "ISO_DUMMY_DATE"
    }
  }
}`

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type EmbedCollUser struct {
	Tasks []Task `json:"tasks" _id_:"user-embed-coll" _hd_:"c"`
}

var DummyEmbedCollUser = EmbedCollUser{
	Tasks: []Task{
		{
			"test task",
			TaskDetails{Date: "ISO_DUMMY_DATE"},
		},
	},
}

const dummyEmbedCollDataJSON = `{
  "tasks": [
    {
      "name": "test task",
      "details": {
        "date": "ISO_DUMMY_DATE"
      }
    }
  ]
}`

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Project test case management setup
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestCase struct {
	ID   string `json:"id" _id_:"test-case"`
	Name string `json:"name" _hd_:"c"`
}

type TestSuite struct {
	ID    string     `json:"id" _id_:"test-suite"`
	Name  string     `json:"name" _hd_:"c"`
	Tests []TestCase `json:"tests" _hd_:"c"`
}

type Project struct {
	ID     string      `json:"id" _id_:"project"`
	Name   string      `json:"name" _hd_:"c"`
	Suites []TestSuite `json:"suites" _hd_:"c"`
}

var DummyProject = Project{
	Name: "p1",
	Suites: []TestSuite{
		{
			Name: "s1",
			Tests: []TestCase{
				{
					Name: "s1t1",
				},
			},
		},
	},
}

var DummyProjectJSON = `{
  "name": "p1",
  "suites": [
    {
      "name": "s1",
      "tests": [
        {
          "name": "s1t1"
        }
      ]
    }
  ]
}`
