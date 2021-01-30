<!--
order: 0
title: Authz Overview
parent:
  title: "authz"
-->

# `authz`

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
    - [MsgGrantAuthorization](03_messages.md#MsgGrantAuthorization)
    - [MsgRevokeAuthorization](03_messages.md#MsgRevokeAuthorization)
    - [MsgExecAuthorized](03_messages.md#MsgExecAuthorized)
4. **[Events](04_events.md)**
    - [Keeper](04_events.md#Keeper)
    
## Abstract
`x/authz` is an implementation of a Cosmos SDK module, per [ADR 30](docs/architecture/adr-030-authz-module.md), that allows 
granting arbitrary privileges from one account (the granter) to another account (the grantee).Authorizations must be granted for a particular Msg service methods one by one using an implementation of `Authorization` interface.