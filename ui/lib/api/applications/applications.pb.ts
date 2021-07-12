/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"

export enum AddApplicationRequestDeploymentType {
  kustomize = "kustomize",
  helm = "helm",
}

export enum AddApplicationRequestSourceType {
  git = "git",
  helm_repo = "helm_repo",
}

export type Application = {
  name?: string
  path?: string
  url?: string
}

export type ListApplicationsRequest = {
  namespace?: string
}

export type ListApplicationsResponse = {
  applications?: Application[]
}

export type GetApplicationRequest = {
  name?: string
  namespace?: string
}

export type GetApplicationResponse = {
  application?: Application
}

export type AddApplicationRequest = {
  name?: string
  url?: string
  path?: string
  branch?: string
  deploymentType?: AddApplicationRequestDeploymentType
  chart?: string
  sourceType?: AddApplicationRequestSourceType
  appConfigUrl?: string
  namespace?: string
  dryRun?: boolean
  autoMerge?: boolean
}

export type AddApplicationResponse = {
  application?: Application
}

export class Applications {
  static ListApplications(req: ListApplicationsRequest, initReq?: fm.InitReq): Promise<ListApplicationsResponse> {
    return fm.fetchReq<ListApplicationsRequest, ListApplicationsResponse>(`/v1/applications?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetApplication(req: GetApplicationRequest, initReq?: fm.InitReq): Promise<GetApplicationResponse> {
    return fm.fetchReq<GetApplicationRequest, GetApplicationResponse>(`/v1/applications/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static AddApplication(req: AddApplicationRequest, initReq?: fm.InitReq): Promise<AddApplicationResponse> {
    return fm.fetchReq<AddApplicationRequest, AddApplicationResponse>(`/v1/applications`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}