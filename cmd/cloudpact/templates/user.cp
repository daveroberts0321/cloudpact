module {{.ProjectName}}Model

define record User
    id: uuid
    name: text
    email: email
    age: number
    createdAt: datetime

