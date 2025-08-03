module {{.ProjectName}}Service

define record User
    name: text
    email: text
    age: number

function validateUser(user: User) returns boolean
    why: "Ensures user data meets basic requirements"
    do:
        if user.age < 18
            then return false
        return true