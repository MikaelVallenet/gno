# daokit

# 1. Introduction

A **Decentralized Autonomous Organization (DAO)** is a self-governing entity that operates through smart contracts, enabling transparent decision-making without centralized control.

`daokit` is a gnolang package for creating complex DAO models. It introduces a new framework based on conditions, composed of :
- `daokit` : Core package for building DAOs, proposals, and actions
- `basedao` : Extension with membership and role management
- `daocond`: Stateless condition engine for evaluating proposals

# 2. What is `daokit` ?

`daokit` provides a powerful condition and role-based system to build flexible and programmable DAOs.

## Key Features:
- Create proposals that include complex execution logic
- Attach rules (conditions) to each resource
- Assign roles to users to structure permissions and governance

## 2.1 Key Concepts

- **Proposal**: A request to execute a **resource**. Proposals are voted on and executed only if predefined **conditions** are met.
- **Resource**: An executable action within the DAO. Each resource is governed by a **condition**.
- **Condition**: A set of rules that determine whether a proposal can be executed.
- **Role**: Labels that assign governance power or permissions to DAO members.

**Example Use Case**: A DAO wants to create a proposal to spend money from its treasury.

**Rules**:
- `SpendMoney` is a resource with a condition requiring:
	- 50% approval from the administration board
	- Approval from the CFO

**Outcome**:
- Any user can propose to spend money
- Only board and CFO votes are considered
- The proposal executes only if the condition is satisfied

# 3. Architecture

DAOkit framework is composed of three packages:

## 3.1 [daocond](../daocond/README.md)

`daocond` provides a stateless condition engine used to evaluate if a proposal should be executed.

### Interface
```go
type Condition interface {
	// checks if the condition is satisfied based on current votes.
	Eval(votes map[string]Vote) bool
	// TODO Check what is this 
	// Signal returns a value from 0.0 to 1.0 to indicate how close the condition is to being met.
	Signal(votes map[string]Vote) float64

	// return a static human-readable representation of the condition.
	Render() string
	// return a dynamic representation with vote context included.
	RenderWithVotes(votes map[string]Vote) string
}
```

### Built-in Conditions
`daocond` provides several built-in conditions to cover common governance scenarios.

```go
// MembersThreshold requires that a specified fraction of all DAO members approve the proposal.
// - threshold: A value between 0.0 and 1.0 representing the minimum approval percentage required.
// - isMemberFn: A function to verify if an ID corresponds to a DAO member.
// - membersCountFn: A function returning the total number of members in the DAO.
func MembersThreshold(threshold float64, isMemberFn func(memberId string) bool, membersCountFn func() uint64) Condition

// RoleThreshold requires that a certain percentage of members holding a specific role approve.
// - threshold: Minimum fraction (0.0 to 1.0) of role members needed to approve.
// - role: The name of the role (e.g., "admin") to check votes against.
// - hasRoleFn: Function to check if a member has the specified role.
// - usersRoleCountFn: Function returning the total number of users with the role.
func RoleThreshold(threshold float64, role string, hasRoleFn func(memberId string, role string) bool, usersRoleCountFn func(role string) uint32) Condition

// RoleCount requires a fixed minimum number of members holding a specific role to approve.
// - count: The minimum number of approving votes from members with the role.
// - role: The role name to consider (e.g., "finance-officer").
// - hasRoleFn: Function to check if a member has the role.
func RoleCount(count uint64, role string, hasRoleFn func(memberId string, role string) bool) Condition
```

### Logical Composition
You can combine multiple conditions to create complex governance rules using logical operators:

```go
// And returns a condition that is satisfied only if *all* provided conditions are met.
func And(conditions ...Condition) Condition
// Or returns a condition that is satisfied if *any* one of the provided conditions is met.
func Or(conditions ...Condition) Condition
```

**Example**:
```go
// Require both admin approval and at least one CFO
cond := daocond.And(
    daocond.RoleThreshold(0.5, "admin", hasRole, roleCount),
    daocond.RoleCount(1, "CFO", hasRole),
)
```

Conditions are stateless for flexibility and scalability.

## 3.2 daokit

`daokit` provides the core mechanics:

### Core Structure:
```go
type Core struct {
	Resources *ResourcesStore
	Proposals *ProposalsStore
}
```

### DAO Interface:
```go
type DAO interface {
	Propose(req ProposalRequest) uint64
	Execute(id uint64)
	Vote(id uint64, vote daocond.Vote)
}
```
> [Code Example of a Basic DAO](#4-code-example-of-a-basic-dao)

### Proposal Lifecycle

Each proposal goes through the following states:

1. **Open**: 
- Initial state after proposal creation.
- Accepts votes from eligible participants.

2. **Passed**
- Proposal has gathered enough valid votes to meet the condition.
- Voting is **closed** and cannot be modified.
- The proposal is now eligible for **execution**.

3. **Executed**
- Proposal action has been successfully carried out.
- Final state — proposal can no longer be voted on or modified.


## 3.3 [basedao](../basedao/README.md)

`basedao` extends `daokit` to handle members and roles management.
It handles who can participate in a DAO and what permissions they have.

### Core Types
```go
type MembersStore struct {
	Roles   *avl.Tree 
	Members *avl.Tree 
}
```

### Initialize the DAO
Create a `MembersStore` structure to initialize the DAO with predefined roles and members.

```go
roles := []basedao.RoleInfo{
	{Name: "admin", Description: "Administrators"},
	{Name: "finance", Description: "Handles treasury"},
}

members := []basedao.Member{
	{Address: "g1abc...", Roles: []string{"admin"}},
	{Address: "g1xyz...", Roles: []string{"finance"}},
}

store := basedao.NewMembersStore(roles, members)
```

### Example Usage
```go
store := basedao.NewMembersStore(nil, nil)

// Add a role and assign it
store.AddRole(basedao.RoleInfo{Name: "moderator", Description: "Can moderate posts"})
store.AddMember("g1alice...", []string{"moderator"})

// Update role assignment
store.AddRoleToMember("g1alice...", "editor")
store.RemoveRoleFromMember("g1alice...", "moderator")

// Inspect the state
fmt.Println("Is Alice a member?", store.IsMember("g1alice..."))
fmt.Println("Is Alice an editor?", store.HasRole("g1alice...", "editor"))
fmt.Println("All Members (JSON):", store.GetMembersJSON())
```

### Creating a DAO:

```go
func New(conf *Config) (daokit.DAO, *DAOPrivate)
```

#### Key Structures:
- `DAOPrivate`: Full access to internal DAO state
- `daokit.DAO`: External interface for DAO interaction


### Configuration:
```go
type Config struct {
	Name              string
	Description       string
	ImageURI          string
	Members           *MembersStore
	NoDefaultHandlers bool
	InitialCondition  daocond.Condition
	SetProfileString  ProfileStringSetter
	GetProfileString  ProfileStringGetter
	NoCreationEvent   bool
}
```

- `MembersStore`: Use `basedao.NewMembersStore(...)` to create members and roles.
- `ProfileStringSetter` / `Getter`: Optional helpers to store profile data (e.g., from `/r/demo/profile`).
- `InitialCondition`: Default rule applied to all built-in DAO actions.
- `NoDefaultHandlers`: Set to `true` to disable built-in actions like add/remove member.
- `NoCreationEvent`: Set to `true` if you don’t want a "DAO Created" event to be emitted.

# 4. Code Example of a Basic DAO

```go
package daokit_demo

import (
	"gno.land/p/samourai/basedao"
	"gno.land/p/samourai/daocond"
	"gno.land/p/samourai/daokit"
	"gno.land/r/demo/profile"
)

var (
	DAO        daokit.DAO // External interface for DAO interaction
	daoPrivate *basedao.DAOPrivate // Full access to internal DAO state
)

func init() {
	initialRoles := []basedao.RoleInfo{
		{Name: "admin", Description: "Admin is the superuser"},
		{Name: "public-relationships", Description: "Responsible of communication with the public"},
		{Name: "finance-officer", Description: "Responsible of funds management"},
	}

	initialMembers := []basedao.Member{
		{Address: "g126...zlg", Roles: []string{"admin", "public-relationships"}},
		{Address: "g1ld6...3jv", Roles: []string{"public-relationships"}},
		{Address: "g1r69...0tth", Roles: []string{"finance-officer"}},
		{Address: "g16jv...6e0r", Roles: []string{}},
	}

	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)

	membersMajority := daocond.MembersThreshold(0.6, memberStore.IsMember, memberStore.MembersCount)
	publicRelationships := daocond.RoleCount(1, "public-relationships", memberStore.HasRole)
	financeOfficer := daocond.RoleCount(1, "finance-officer", memberStore.HasRole)

	// `and` and `or` use va_args so you can pass as many conditions as needed
	adminCond := daocond.And(membersMajority, publicRelationships, financeOfficer)

	DAO, daoPrivate = basedao.New(&basedao.Config{
		Name:             "Demo DAOKIT DAO",
		Description:      "This is a demo DAO built with DAOKIT",
		Members:          memberStore,
		InitialCondition: adminCond,
		GetProfileString: profile.GetStringField,
		SetProfileString: profile.SetStringField,
	})
}

func Vote(proposalID uint64, vote daocond.Vote) {
	DAO.Vote(proposalID, vote)
}

func Execute(proposalID uint64) {
	DAO.Execute(proposalID)
}

func Render(path string) string {
	return daoPrivate.Render(path)
}
```

# 5. Create Custom Resources

To add new behavior to your DAO — or to enable others to integrate your package into their own DAOs — define custom resources by implementing:

```go
type Action interface {
	Type() string // TODO Type of the action (like a slug) // human readable ?
	String() string // return human-readable content of the action // TODO explain the payload to string
}

type ActionHandler interface {
	Type() string // TODO // return the type of the action (like a slug)
	Execute(action Action) // Executes logic associated with the action
}
```
This allows DAOs to execute arbitrary logic or interact with Gno packages through governance-approved decisions.

``daokit`` provide a generic implementation of ``Action`` and ``ActionHandler``, available at the [``./actions.gno``](./actions.gno) file.

// TODO Add an example of implementation

## Steps to Add a Custom Resource:
1. Define the path of the action, it should be unique 
```go
// XXX: pkg "/p/samourai/blog" - does not exist, it's just an example
const ActionNewPostKind = "gno.land/p/samourai/blog.NewPost"
```

2. Create the structure type of the payload
```go
type ActionNewPost struct {
	Title string
	Content string
}
```

3. Implement the action and handler
```go
func NewPostAction(title, content string) daokit.Action {
	// def: daoKit.NewAction(kind: String, payload: interface{})
	return daokit.NewAction(ActionNewPostKind, &ActionNewPost{
		Title:   title,
		Content: content,
	})
}

func NewPostHandler(blog *Blog) daokit.ActionHandler {
	// def: daoKit.NewActionHandler(kind: String, payload: func(interface{}))
	return daokit.NewActionHandler(ActionNewPostKind, func(payload interface{}) {
		action, ok := payload.(*ActionNewPost)
		if !ok {
			panic(errors.New("invalid action type"))
		}
		blog.NewPost(action.Title, action.Content)
	})
}
```

4. Register the resource
```go
resource := daokit.Resource{
    Condition: daocond.NewRoleCount(1, "CEO", daoPrivate.Members.HasRole),
    Handler: blog.NewPostHandler(blog),
}
daoPrivate.Core.Resources.Set(&resource)
```
