# CloudPact Grammar Guide

A comprehensive reference for the CloudPact language syntax, designed for human-AI collaborative programming.

## Table of Contents

1. [Language Overview](#language-overview)
2. [Current Implementation Status](#current-implementation-status)
3. [Basic Syntax Elements](#basic-syntax-elements)
4. [Data Types](#data-types)
5. [Module Structure](#module-structure)
6. [Model Definitions](#model-definitions)
7. [Function Definitions](#function-definitions)
8. [Control Flow](#control-flow)
9. [AI Integration Syntax](#ai-integration-syntax)
10. [Semantic Types](#semantic-types)
11. [Examples](#examples)

## Language Overview

CloudPact is a domain-specific language designed to bridge human intent and AI understanding. It compiles to both Go and TypeScript from a single source, with built-in business context preservation.

### Design Principles
- **Human Readable**: Natural language-like syntax
- **AI Friendly**: Explicit business context through `why:` clauses
- **Target Agnostic**: Single source compiles to multiple targets
- **Context Preserving**: Business logic explanation survives compilation

## Current Implementation Status

### ✅ Implemented (v0.1.0)
```ebnf
File        := { Declaration }
Declaration := Model | Function | Assignment
Model       := 'model' IDENT '{' { Field } '}'
Field       := IDENT ':' Type [ RelationshipDecl ]
Type        := IDENT [ TypeConstraints ]
Function    := 'function' IDENT '(' ParamList ')' [ 'returns' Type ] FunctionBody
FunctionBody:= WhyClause 'do:' StatementList
WhyClause   := 'why:' STRING
Assignment  := 'assign-use' IDENT 'as' Type [ 'why:' STRING ] [ 'validate:' ValidationRule ]
RelationshipDecl := 'belongs_to' | 'has_one' | 'has_many' | 'references'
```

### 🚧 In Development (Target Syntax)
```ebnf
File           := ModuleDecl { Declaration }
ModuleDecl     := 'module' IDENT
Declaration    := RecordDef | FunctionDef | TypeDef
RecordDef      := 'define' 'record' IDENT { FieldDef }
FieldDef       := IDENT ':' Type
FunctionDef    := 'function' IDENT '(' ParamList ')' [ 'returns' Type ] WhyClause DoBlock
DoBlock        := 'do:' { Statement }
Statement      := IfStatement | Assignment | Return | Expression
```

## Basic Syntax Elements

### Identifiers
- Must start with letter or underscore
- Can contain letters, numbers, underscores
- Case sensitive

### Comments
```cloudpact
// Single line comment
/* Multi-line
   comment */
```

### Literals
```cloudpact
// String literals
"Hello, World!"
'Single quotes also work'

// Number literals
42          // Integer
3.14        // Float
1_000_000   // Underscores for readability

// Boolean literals
true
false
```

## Data Types

### Basic Types
```cloudpact
text        // String data
number      // Numeric data (int or float)
boolean     // true/false values
```

### Semantic Types
CloudPact includes rich semantic types for better validation and documentation:

#### Identity Types
```cloudpact
email           // Email addresses with validation
phone           // Phone numbers
uuid            // UUID identifiers
username        // User identifiers
```

#### Location Types
```cloudpact
address         // Street addresses
zip_code        // Postal codes
country_code    // ISO country codes (US, CA, etc.)
state_code      // State/province codes
```

#### Financial Types
```cloudpact
usd_currency    // US Dollar amounts
eur_currency    // Euro amounts
percentage      // Percentage values (0-100)
```

#### Temporal Types
```cloudpact
date            // Date values (YYYY-MM-DD)
datetime        // Date and time (ISO 8601)
time            // Time values (HH:MM:SS)
duration        // Time durations (ISO 8601)
```

#### Security Types
```cloudpact
password        // Password fields (masked)
token           // Authentication tokens
api_key         // API keys
```

## Module Structure

### Current Implementation
```cloudpact
// Multiple models in a file
model User {
    id: String
    name: String
    email: String
}

model Order {
    id: String
    userId: String belongs_to User
    total: Float
}
```

### Target Implementation
```cloudpact
module UserService

define record User
    id: uuid
    name: text
    email: email
    createdAt: datetime

define record Order
    id: uuid
    userId: uuid
    total: usd_currency
    status: text
```

## Model Definitions

### Current Syntax
```cloudpact
model ModelName {
    fieldName: FieldType
    relationField: TargetType relationship_type TargetModel
}
```

### Relationships
```cloudpact
model User {
    id: String
    profile: Profile has_one Profile
    orders: Order has_many Order
}

model Profile {
    userId: String belongs_to User
    bio: String
}

model Order {
    userId: String references User
    total: Float
}
```

### Target Syntax
```cloudpact
define record User
    id: uuid
    name: text
    email: email
    age: number
    profile: Profile         // Type references
    orders: list[Order]      // Collection types
```

## Function Definitions

### Current Implementation
```cloudpact
function functionName(param1: Type1, param2: Type2) returns ReturnType
    why: "Business explanation of what this function does"
    do:
        // Simple statement parsing (limited)
```

### Target Implementation
```cloudpact
function validateUser(user: User) returns boolean
    why: "Ensures user data meets business requirements for registration"
    do:
        if user.age < 18
            then return false
        if not isValidEmail(user.email)
            then return false
        return true

function registerUser(name: text, email: email, password: text) returns User
    why: "Creates a new user account with proper validation and security"
    do:
        if not validateEmail(email)
            then fail "Invalid email address"
        set hashedPassword = hashPassword(password)
        create user with:
            name = name
            email = email
            password = hashedPassword
            createdAt = now()
        return user
```

## Control Flow

### Conditional Statements
```cloudpact
if condition
    then statement
else if otherCondition
    then otherStatement
else
    defaultStatement
```

### Pattern Matching (Planned)
```cloudpact
match user.status:
    case "active" then processActiveUser(user)
    case "pending" then sendActivationEmail(user)
    case "suspended" then logSuspensionEvent(user)
    default then handleUnknownStatus(user)
```

### Loops (Planned)
```cloudpact
for user in users:
    processUser(user)

while hasMoreData():
    processNextBatch()
```

## AI Integration Syntax

### AI Feedback Annotations
```cloudpact
function calculateShipping(weight: number, zone: number) returns number
    why: "Calculates shipping cost based on weight and destination zone"
    ai-feedback: "Consider adding validation for negative weights"
    ai-suggests: "Cache zone multipliers for better performance"
    ai-security: "No security concerns for this calculation"
    do:
        if weight <= 1
            then base = 5
        return base * zone
```

### AI Decision History
```cloudpact
function hashPassword(input: text) returns text
    why: "Hashes user passwords for secure storage"
    ai-decision-accepted: "Switched from SHA-256 to bcrypt per AI suggestion"
    ai-decision-rejected: "Rejected AI suggestion to add salt parameter - handled internally"
    do:
        use bcrypt algorithm
        return hashed result of input
```

## Semantic Types

### Type Constraints
Each semantic type includes built-in validation rules:

```cloudpact
define record User
    email: email              // Validates email format
    phone: phone              // Validates phone number format
    age: number               // Basic number validation
    salary: usd_currency      // Validates currency format, >= 0
    website: url              // Validates URL format
    country: country_code     // Validates ISO country codes
```

### Custom Validation (Planned)
```cloudpact
define type CustomEmail as email
    validate: domain must be "company.com"
    why: "Only company email addresses allowed"

define type ProductPrice as usd_currency
    validate: value between 0.01 and 9999.99
    why: "Product prices must be reasonable range"
```

## Examples

### Complete User Service
```cloudpact
module UserService

define record User
    id: uuid
    name: text
    email: email
    password: password
    age: number
    createdAt: datetime

function validateAge(age: number) returns boolean
    why: "Ensures users meet minimum age requirement"
    do:
        return age >= 18

function validateEmail(email: email) returns boolean
    why: "Checks email format and domain restrictions"
    do:
        if not email contains "@"
            then return false
        return email endsWith ".com" or email endsWith ".org"

function hashPassword(password: text) returns text
    why: "Securely hashes passwords using bcrypt"
    do:
        use bcrypt algorithm with salt rounds 12
        return hashed result of password

function createUser(name: text, email: email, password: text, age: number) returns User
    why: "Creates a new user with full validation and security measures"
    do:
        if not validateAge(age)
            then fail "User must be at least 18 years old"
        if not validateEmail(email)
            then fail "Invalid email address"
        set hashedPassword = hashPassword(password)
        create user with:
            id = generateUUID()
            name = name
            email = email
            password = hashedPassword
            age = age
            createdAt = now()
        return user
```

### E-commerce Order System
```cloudpact
module OrderService

define record Product
    id: uuid
    name: text
    price: usd_currency
    inStock: boolean

define record OrderItem
    productId: uuid
    quantity: number
    price: usd_currency

define record Order
    id: uuid
    userId: uuid
    items: list[OrderItem]
    total: usd_currency
    status: text
    createdAt: datetime

function calculateOrderTotal(items: list[OrderItem]) returns usd_currency
    why: "Calculates total cost including tax and shipping"
    do:
        set subtotal = 0
        for item in items:
            subtotal = subtotal + (item.price * item.quantity)
        set tax = subtotal * 0.08
        set shipping = calculateShipping(subtotal)
        return subtotal + tax + shipping

function validateOrder(order: Order) returns boolean
    why: "Ensures order meets business rules before processing"
    do:
        if order.items is empty
            then return false
        if order.total <= 0
            then return false
        return true
```

## Compilation Targets

### Go Output
```go
// Generated from CloudPact
type User struct {
    ID        string    `json:"id" validate:"required,uuid"`
    Name      string    `json:"name" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Age       int       `json:"age" validate:"required,min=18"`
    CreatedAt time.Time `json:"created_at"`
}

func ValidateAge(age int) bool {
    // Business logic: Ensures users meet minimum age requirement
    return age >= 18
}
```

### TypeScript Output
```typescript
// Generated from CloudPact
export interface User {
  id: string;          // UUID format
  name: string;
  email: string;       // Email format validation
  age: number;         // Minimum 18
  createdAt: string;   // ISO datetime
}

export function validateAge(age: number): boolean {
  // Business logic: Ensures users meet minimum age requirement
  return age >= 18;
}
```

## Parser Implementation Notes

### Current Limitations
- Module declarations not yet supported
- Limited function body parsing
- Control flow parsing incomplete
- AI annotation parsing not implemented

### Next Implementation Steps
1. Add `module` keyword support
2. Implement `define record` syntax
3. Enhance function body parsing for control flow
4. Add AI annotation parsing
5. Implement semantic type validation generation

---

*This grammar guide will be updated as CloudPact evolves. The goal is to maintain backward compatibility while adding new collaborative features.*