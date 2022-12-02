package models

type Molt struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	ID           string `dynamodbav:"id"`
	Created      string `dynamodbav:"created"`
	Author       string `dynamodbav:"author"`
	Content      string `dynamodbav:"content"`
	Deleted      bool   `dynamodbav:"deleted"`
	LikeCount    int    `dynamodbav:"like_count"`
	RemoltCount  int    `dynamodbav:"remolt_count"`
	CommentCount int    `dynamodbav:"comment_count"`
}

func (m *Molt) ById(svc ItemService, tablename string, text string) {

}

func (m *Molt) ByAuthor(svc ItemService, tablename string, text string) {

}

func (m *Molt) ByTime(svc ItemService, tablename string, text string) {

}

func (m *Molt) Create(svc ItemService, tablename string, text string) {

}

func (m *Molt) Re(svc ItemService, tablename string, text string) {

}

func (m *Molt) Delete(svc ItemService, tablename string, text string) {

}

func (m *Molt) Edit(svc ItemService, tablename string, text string) {

}
