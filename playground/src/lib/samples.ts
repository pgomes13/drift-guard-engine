import { SchemaType } from "./types";

export interface Sample {
  base: string;
  head: string;
}

export const SAMPLES: Record<SchemaType, Sample> = {
  openapi: {
    base: `openapi: "3.0.0"
info:
  title: Users API
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                  total:
                    type: integer
    post:
      operationId: createUser
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - email
              properties:
                name:
                  type: string
                email:
                  type: string
      responses:
        "201":
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
    delete:
      operationId: deleteUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "204": {}`,
    head: `openapi: "3.0.0"
info:
  title: Users API
  version: "2.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                  # 'total' removed
    post:
      operationId: createUser
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - email
              properties:
                name:
                  type: string
                email:
                  type: string
                role:
                  type: string
      responses:
        "201":
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
    # DELETE /users/{id} removed
  /posts:
    get:
      operationId: listPosts
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array`,
  },

  graphql: {
    base: `type Query {
  user(id: ID!): User
  users(limit: Int, offset: Int): [User!]!
  search(query: String!): SearchResult
}

type Mutation {
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
  deleteUser(id: ID!): Boolean!
}

type User {
  id: ID!
  email: String!
  name: String
  role: UserRole!
  address: Address
}

type Address {
  street: String!
  city: String!
  country: String!
}

enum UserRole {
  ADMIN
  VIEWER
  EDITOR
}

input CreateUserInput {
  email: String!
  name: String
  role: UserRole
}

input UpdateUserInput {
  name: String
  role: UserRole
}

union SearchResult = User | Address

interface Node {
  id: ID!
}`,
    head: `type Query {
  user(id: ID!, includeDeleted: Boolean): User
  users(limit: Int, offset: Int): [User!]!
  search(query: String!, type: String): SearchResult
}

type Mutation {
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
  # deleteUser removed
}

type User {
  id: ID!
  email: String!
  name: String
  role: UserRole!
  # address removed
  createdAt: String
}

type Address {
  street: String!
  city: String!
  country: String!
  postcode: String
}

enum UserRole {
  ADMIN
  VIEWER
  # EDITOR removed
}

input CreateUserInput {
  email: String!
  name: String
  role: UserRole!
}

input UpdateUserInput {
  name: String
  role: UserRole
  notes: String
}

type Post {
  id: ID!
  title: String!
}

union SearchResult = User | Address | Post

interface Node {
  id: ID!
}`,
  },

  grpc: {
    base: `syntax = "proto3";

package users.v1;

service UserService {
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  rpc ListUsers (ListUsersRequest) returns (stream ListUsersResponse);
  rpc DeleteUser (DeleteUserRequest) returns (DeleteUserResponse);
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string id    = 1;
  string name  = 2;
  string email = 3;
}

message ListUsersRequest {
  int32 page      = 1;
  int32 page_size = 2;
}

message ListUsersResponse {
  repeated GetUserResponse users = 1;
  int32 total                   = 2;
}

message DeleteUserRequest {
  string id = 1;
}

message DeleteUserResponse {
  bool success = 1;
}`,
    head: `syntax = "proto3";

package users.v1;

service UserService {
  rpc GetUser (GetUserRequestV2) returns (GetUserResponse);
  rpc ListUsers (stream ListUsersRequest) returns (stream ListUsersResponse);
  // DeleteUser removed
  rpc CreateUser (CreateUserRequest) returns (CreateUserResponse);
}

service AdminService {
  rpc BanUser (BanUserRequest) returns (BanUserResponse);
}

message GetUserRequestV2 {
  string id     = 1;
  string locale = 2;
}

message GetUserResponse {
  string id   = 1;
  string name = 2;
  // email removed
  string role = 4;
}

message ListUsersRequest {
  int32 page      = 1;
  int32 page_size = 2;
}

message ListUsersResponse {
  repeated GetUserResponse users = 1;
  // total removed
}

message CreateUserRequest {
  string name  = 1;
  string email = 2;
}

message CreateUserResponse {
  string id = 1;
}

message BanUserRequest {
  string user_id = 1;
}

message BanUserResponse {
  bool success = 1;
}`,
  },
};

export const MONACO_LANGUAGE: Record<SchemaType, string> = {
  openapi: "yaml",
  graphql: "graphql",
  grpc: "protobuf",
};
