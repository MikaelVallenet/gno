# 1. Introduction

A **Decentralized Autonomous Organization (DAO)** is a self-governing entity that operates through smart contracts, enabling transparent decision-making without centralized control.

Daokit is a gnolang package for creating complex DAO models. It introduces a new framework based on conditions. BaseDAO is an extension of DAOkit that handles the members & roles.

# 2. Why daokit ?

The advantages:
- Condition system to create complex proposals combining roles, counting, percentages, logical operations, etc.
- Role system to represent the organization & responsabilities of the members

## 2.1 Key Concepts

To understand how it works, here are three fundamental concepts:

- **Proposal:** A proposal is a request to execute a resource (an action) as the organization itself. A proposal is created by a member and other members can vote on it. A proposal use the condition attached the resource to evaluate if the proposal can be executed. A proposal can have different states like: open, executed, closed.
- **Resource:** An executable action within the DAO, triggered through the **proposal system**. A resource also contains a condition that must be met before the proposal can be executed.
- **Condition:** A set of **rules attached to a resource**, which must be met before the corresponding proposal can be executed.

A central concept but not fundamental is the **Role**. A role is a label assigned to a user (or entity) that grants specific **permissions and governance power** based on the organization’s structure. you can create organization without roles, but using it in combination with daocond will allow you to create more complex proposals.

To give an example of how this key concepts interacts, let's consider a DAO that wants to create a proposal to **spend money** from the DAO's treasury.

The DAO's constitution may include the following rules:

- SpendMoney is a resource that allows to spend money from the DAO's treasury.
- To execute a resource a proposal must be created
- A condition is attached to the SpendMoney resource, let's say for the condition to be met the proposal must be supported by 50% of the administration board, and by the Chief Financial Officer (CFO)

In this case:
- Any user can create a proposal to spend money
- Only the votes of the administration board & the CFO will matters
- The proposal will only be executed if it gets the majority of the votes of the administration board & the CFO

# 3. Architecture

the daokit framework is based on the following packages:

## 3.1 daocond

``daocond`` is a package that allows to create conditions (treshold, count, etc) can be combined with roles but is package agnostic and takes function as parameter.

Here is the ``Condition`` interface:

```go

type Condition interface {
	Eval(votes map[string]Vote) bool
	Signal(votes map[string]Vote) float64

	Render() string
	RenderWithVotes(votes map[string]Vote) string
}
```

A condition does not handle the voting process & storage, it only evaluate if the condition is met based on the votes.

We have conditions for members treshold, roles treshold, roles count ...etc

```go
func MembersThreshold(threshold float64, isMemberFn func(memberId string) bool, membersCountFn func() uint64) Condition
func RoleThreshold(threshold float64, role string, hasRoleFn func(memberId string, role string) bool, usersRoleCountFn func(role string) uint32) Condition
func RoleCount(count uint64, role string, hasRoleFn func(memberId string, role string) bool) Condition
```

We also have conditions that can be used to build more complex conditions using logical operators like AND & OR.

```go
func And(conditions ...Condition) Condition
func Or(conditions ...Condition) Condition
```

But feel free to create your own conditions based on your own needs & external resources.

We decided to make our conditions stateless first because adding state add some drawbacks depending on the structure (size, and kind of proposals) of the DAO.

## 3.2 daokit

``daokit`` is a package that creates the core of the DAO (resources, proposals). It imports ``daocond`` to attach it to the resource structure.

Here is the ``Core`` structure:

```go
type Core struct {
	Resources *ResourcesStore
	Proposals *ProposalsStore
}
```

``daokit`` package also provide an interface for the public usage of the DAO:

```go
type DAO interface {
	Propose(req ProposalRequest) uint64
	Execute(id uint64)
	Vote(id uint64, vote daocond.Vote)
}
```

``daokit`` also provide an ``Action`` and ``ActionHandler`` interface used to create new resources and handlers for DAOs, discover more in the [5. Create Custom Resources](#5-create-custom-resources) section.

## 3.3 basedao

``basedao`` is a wrapper around daokit that handles the memberstore. The memberstore is a structure that includes few methods to handle the members and the roles. It also provider the rendering and create some basic resources like add/remove member, add/remove role, etc.

``basedao`` split the DAO in two structures, the ``DAOPrivate`` and the ``daoPublic``.
The difference is that the ``DAOPrivate`` is meant to be used internally by the DAO and not exposed to the outside. By accessing the ``DAOPrivate`` you could modify the DAO core, the memberstore or everything in the DAO. The ``daoPublic`` is the public part of the DAO, it's the part that will be exposed to the outside.
It's an implementation of [``DAO`` interface](#32-daokit) from ``daokit`` package.

To create a new DAO the endpoint is the following:

```go
func New(conf *Config) (daokit.DAO, *DAOPrivate)
```

The ``Config`` structure is the following:

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

- The ``MembersStore`` can be created with ``basedao.NewMembersStore(...)``
- The ``ProfileStringSetter`` and ``ProfileStringGetter`` are just here to set and get the profile of the DAO. could be the functions from the ``/r/demo/profile`` package for example.
- The ``InitialCondition`` is meant to be attached to the default resources of the DAO.
- The ``NoDefaultHandlers`` is just here to enable/disable the default handlers of the DAO.
- ``NoCreationEvent`` will disable the event emitted when the DAO is created.

# 4. Code example of a basic DAO

```go
package daokit_demo

import (
	"gno.land/p/samourai/basedao"
	"gno.land/p/samourai/daocond"
	"gno.land/p/samourai/daokit"
	"gno.land/r/demo/profile"
)

var (
	DAO        daokit.DAO // exposed to the outside
	daoPrivate *basedao.DAOPrivate // internal use
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

	// create the member store now to be able to use it in the condition
	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)

	membersMajority := daocond.MembersThreshold(0.6, memberStore.IsMember, memberStore.MembersCount)
	publicRelationships := daocond.RoleCount(1, "public-relationships", memberStore.HasRole)
	financeOfficer := daocond.RoleCount(1, "finance-officer", memberStore.HasRole)

	// and & or use va_args so you can pass as many conditions as you want
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

There is a good chance that you will need to create custom resources for your DAO to met your need. You could also want to create custom resources for your package without being yourself a daokit consumer but to give people the ability to use your package in their DAOs.

To do so, you will need to implement the ``Action`` and ``ActionHandler`` interface.

Here is the interfaces:

```go
type Action interface {
	String() string // return a string representation of the action
	Type() string // return the type of the action (like a slug)
}

type ActionHandler interface {
	Execute(action Action) // execute the action
	Type() string // return the type of the action (like a slug)
}
```

The ``daokit`` package provide a generic implementation of the ``Action`` and ``ActionHandler`` interface to allows users to create in a simple way their own resources. You can take a look at the [``./actions.gno``](./actions.gno) file to see how the generic implementation works.


So to create your own resource follow these steps:

```go
//XXX: pkg /p/samourai/blog (this package does not exist, it's just an example)
// Think about crossing if you write the actions within realm and not package


// 1. Define the kind (slug) of the action, it should be unique 
const ActionNewPostKind = "gno.land/p/samourai/blog.NewPost"

// 2. Define the action structure (the payload of the action)
// It should contain the data needed to execute the action
type ActionNewPost struct {
	Title string
	Content string
}

// 3. Use the generic action handler to create the factory of handlers & actions
func NewPostHandler(blog *Blog) daokit.ActionHandler {
	return daokit.NewActionHandler(ActionNewPostKind, func(payload interface{}) {
		action, ok := payload.(*ActionNewPost)
		if !ok {
			panic(errors.New("invalid action type"))
		}
		blog.NewPost(action.Title, action.Content)
	})
}

func NewPostAction(title, content string) daokit.Action {
	return daokit.NewAction(ActionNewPostKind, &ActionNewPost{
		Title:   title,
		Content: content,
	})
}

---

// 4. Register the handler to the DAO
// XXX: /r/samourai/samdao

// 4.1 Import the package
import (
    "gno.land/p/samorai/blog"
    "gno.land/p/samourai/daokit"
)

var blog *blog.Blog
// 4.2 Register your resources
func init() {
    ...
    blog = blog.NewBlog()

    resource := daokit.Resource{
        Condition: daocond.NewRoleCount(1, "CEO", daoPrivate.Members.HasRole),
        Handler: blog.NewPostHandler(blog),
    }
    daoPrivate.Core.Resources.Set(&resource)
}

```

Now as the author of the package blog, my users can easily use the blog package within their DAO without having to implement the ``Action`` and ``ActionHandler`` interface.
