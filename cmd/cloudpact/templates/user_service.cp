module {{.ProjectName}}Service

function validateUser(user: User) returns boolean
    ai-feedback: "Consider adding additional validation"
    why: "Ensures user data meets basic requirements"
    do:
        if user.age < 18
            then return false
        return true

