/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum SourceType {
  Git = "Git",
  Bucket = "Bucket",
  Helm = "Helm",
  Chart = "Chart",
}

export type Artifact = {
  checksum?: string
  lastupdateat?: number
  path?: string
  revision?: string
  url?: string
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}

export type GitRepositoryRef = {
  branch?: string
  tag?: string
  semver?: string
  commit?: string
}

export type Source = {
  name?: string
  url?: string
  reference?: GitRepositoryRef
  type?: SourceType
  provider?: string
  bucketname?: string
  region?: string
  namespace?: string
  gitimplementation?: string
  timeout?: string
  secretRefName?: string
  conditions?: Condition[]
  artifact?: Artifact
}