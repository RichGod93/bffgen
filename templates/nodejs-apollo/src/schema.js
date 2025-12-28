/**
 * GraphQL Schema using Apollo Server 4
 * Project: {{PROJECT_NAME}}
 */

export const typeDefs = `#graphql
  type Query {
    """
    Health check endpoint
    """
    health: HealthStatus!
    
    {{ range .BackendServices }}
    """
    Query {{ .Name }} service
    """
    {{ .Name | ToCamelCase }}(id: ID!): {{ .Name | ToPascalCase }}
    {{ .Name | ToCamelCase }}List: [{{ .Name | ToPascalCase }}!]!
    {{ end }}
  }

  type Mutation {
    {{ range .BackendServices }}
    """
    Create new {{ .Name }} entry
    """
    create{{ .Name | ToPascalCase }}(input: {{ .Name | ToPascalCase }}Input!): {{ .Name | ToPascalCase }}!
    
    """
    Update {{ .Name }} entry
    """
    update{{ .Name | ToPascalCase }}(id: ID!, input: {{ .Name | ToPascalCase }}Input!): {{ .Name | ToPascalCase }}!
    
    """
    Delete {{ .Name }} entry
    """
    delete{{ .Name | ToPascalCase }}(id: ID!): Boolean!
    {{ end }}
  }

  type HealthStatus {
    status: String!
    services: [ServiceHealth!]!
  }

  type ServiceHealth {
    name: String!
    status: String!
    url: String!
  }

  {{ range .BackendServices }}
  type {{ .Name | ToPascalCase }} {
    id: ID!
    # TODO: Add your {{ .Name }} fields here
  }

  input {{ .Name | ToPascalCase }}Input {
    # TODO: Add your {{ .Name }} input fields here
  }
  {{ end }}

  # Custom scalars
  scalar DateTime
  scalar JSON
`;
