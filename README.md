# Deprecation Notice

We no longer support this project and will be archiving this repository to reflect the same.
In case needed, the binary can be built using the commands in Makefile

# AWS ADFS login

Library for user login (client side) using AWS ADFS (Active Directory Federation Service).

## Example

Errors are ignored to make example shorter and more readable

```
// Load aws roles
roles, _ := LoadAWSRoles(adfsHost, user, password)

// List all accounts
accounts := roles.Accounts()

// Filter roles by account
accountRoles := roles.RolesByAccountId("123456789")

// Get specific role and log in
admin, _ := roles.RoleByRoleArn("arn:aws:iam::123456789:role/Admin")
creds, _ := admin.Login()

```

MFA Duo

```
devices, _ := MFA(adfsHost, user, password)

// factor can be 'Phone Call', 'Duo Push', or 'Passcode'
// passcode is required only with 'Passcode' factor
roles, _ := devices["phone1"].Factors["Duo Push"].LoadAWSRoles("")
```

# Legal
This project is available under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0.html).

Copyright 2018 Expedia Inc.
