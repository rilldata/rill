---
title: "Secure: Dashboards"
description: Granular access policies for dashboards 
sidebar_label: "Secure: Dashboards"
sidebar_position: 30
---

:::caution

This is an experimental feature.

:::

Rill supports granular access policies for dashboards. You can define access policies for dashboards in the respective `dashboard_name.yaml` file under `policies` section.
See policies configuration reference for all the possible options here.

## Dashboard Policies
Currently, Rill supports the following use cases for dashboards:
- Dashboard level access - Based on this condition you can restrict access to a dashboard for specific users based on user attributes.
- Row level filtering - Based on this condition you can restrict access to certain rows for a specific users based on user attributes.
- Column level access - Based on this condition you can restrict access to certain columns in a dashboard for a specific users based on user attributes. 
Currently only dimensions defined in the dashboard definition can be used for column level access.

### User Attributes
Following user attributes are available for defining policies:
- `name` - name of the user. Example - `John Doe`    
- `email` - email of the user. Example - `john.doe@test.com`
- `domain` - domain of the email. Example - `test.com`
- `groups` - _**list**_ of groups the user belongs to. Currently, we don't support custom user groups, so the only group available is `all`. Example - `all`
- `admin` - whether the user is an admin or not. Example - `true`

User attributes will be namespaced with `.user` prefix while being used, more in the example below
 
## Examples

1. Let's say you want to restrict access to a dashboard only to admin users. You can define the following policy in the `dashboard_name.yaml` file under `policy` section.
    ```yaml 
    policy:
      has_access: "'{{ .user.admin }}' == true"
    ```
   > **_Note:_**  If `policy` section is defined and `has_access` is not, then it will resolve to `false` meaning dashboard won't be accessible to anyone.
2. Along with the above policy, you also want to filter all counts/measures as per the user domain. Assuming you have a `domain` dimension in your model. You can define the following policy in the `dashboard_name.yaml` file under `policy` section.
    ```yaml
    policy:
      has_access: "'{{ .user.admin }}' == true"
      filter: "domain = '{{ .user.domain }}'"
     ```
   > **_Note:_**  `filter` condition needs to be a valid SQL `WHERE` clause. It will be injected in the queries after resolving the user attributes templating.
3. Along with the above policy, you want to exclude certain dimension for user not having `test.com` domain. Assuming you have a `ssn` dimension in your model. You can define the following policy in the `dashboard_name.yaml` file under `policy` section.
    ```yaml
    policy:
      has_access: "'{{ .user.admin }}' == true"
      filter: "domain = '{{ .user.domain }}'"
      exclude:
        - name: ssn
          if: "{{ .user.domain }} != 'test.com'"
    ```
   As per above ssn dimension will be excluded but all other dimensions will be included.
4. Here's one example for column inclusion.

    ```yaml
    policy:
      has_access: "true"
      filter: "domain = '{{ .user.domain }}'"
      include:
        - name: ssn
          if: "'{{ .user.admin }}' == true"
    ```
   > **_Note:_** If include is defined all other dimensions will be excluded. Only one of the
     `include`/`exclude` can be defined at a time.

> **_Note:_** `has_access` and `if` conditions on `include`/`exclude` fields needs to be valid expressions that evaluates to true/false.

## Advanced Examples
Policies are resolved in two phases - 
1. The templating engine replaces placeholder like this `{{ }}` with actual values
2. For `has_access` and `if` conditions, the resultant expression is evaluated to true/false, whereas for `filter` condition the resultant expression is injected in the query.

> **_Note:_** We use sprig templating engine for templating and govaluate for expression evaluation. So any valid operator/expression they support will work here. You can refer to their documentation for more details.
[Here](http://masterminds.github.io/sprig/) are the supported sprig functions and [here](https://github.com/Knetic/govaluate/blob/master/MANUAL.md#operators) are the supported expression evaluation operators. 

Now lets look at some advanced examples.

1. Let's say you want to restrict access to a dashboard only to admin users or users belonging to `test.com` domain.
    ```yaml 
    policy:
      has_access: "'{{ .user.admin }}' == true || '{{ .user.domain }}' == 'test.com'"
    ```
2. Let's say additionally we want to filter queries based on user's groups and there exist a `group` dimension in the model.
    ```yaml 
    policy:
      has_access: "'{{ .user.admin }}' == true || '{{ .user.domain }}' == 'test.com'"
      filter: "groups IN ('{{ .user.groups | join "', '" }}')"
    ```

### Custom attributes
Sometimes you may want to use custom attributes for defining policies other than existing user attributes. For example, you may have a `department` dimension and want to filter the dashboard view as per the users department.
You need to define the mapping of users to department as a separate [source](../reference/project-files/sources) in Rill. 
Then you can refer to the source mapping in the policy. 

For example, lets say we create a file named `mappings.csv` having the following data and put it in `data/` directory.
```csv
email,department
john.doe@test.com,marketing
jane.doe@test.com,tech
```
Now we can define a new source in `sources/mappings.yaml` file as follows:
```yaml
type: "local_file"
path: "data/mappings.csv"
``` 

Now we can refer to this source in the policy as follows:
```yaml
policy:
has_access: "'{{ .user.admin }}' == true || '{{ .user.domain }}' == 'test.com'"
filter: "department IN (SELECT department FROM mappings WHERE email = '{{ .user.email }}' )"
```
