/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../applications/fetch.pb"
export type App = {
  name?: string
  description?: string
  displayName?: string
  id?: string
}

export type AddAppRequest = {
  repoName?: string
  name?: string
  description?: string
  displayName?: string
}

export type AddAppResponse = {
  success?: boolean
  app?: App
}

export type GetAppRequest = {
  repoName?: string
  appName?: string
}

export type GetAppResponse = {
  app?: App
}

export type ListAppRequest = {
  repoName?: string
}

export type ListAppResponse = {
  apps?: App[]
}

export type RemoveAppRequest = {
  name?: string
  namespace?: string
  autoMerge?: boolean
}

export type RemoveAppResponse = {
  success?: boolean
}

export class Apps {
  static AddApp(req: AddAppRequest, initReq?: fm.InitReq): Promise<AddAppResponse> {
    return fm.fetchReq<AddAppRequest, AddAppResponse>(`/v1/app`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetApp(req: GetAppRequest, initReq?: fm.InitReq): Promise<GetAppResponse> {
    return fm.fetchReq<GetAppRequest, GetAppResponse>(`/v1/repo/${req["repoName"]}/app/${req["appName"]}?${fm.renderURLSearchParams(req, ["repoName", "appName"])}`, {...initReq, method: "GET"})
  }
  static ListApps(req: ListAppRequest, initReq?: fm.InitReq): Promise<ListAppResponse> {
    return fm.fetchReq<ListAppRequest, ListAppResponse>(`/v1/repo/${req["repoName"]}/app?${fm.renderURLSearchParams(req, ["repoName"])}`, {...initReq, method: "GET"})
  }
  static RemoveApp(req: RemoveAppRequest, initReq?: fm.InitReq): Promise<RemoveAppResponse> {
    return fm.fetchReq<RemoveAppRequest, RemoveAppResponse>(`/v1/app/${req["name"]}`, {...initReq, method: "DELETE", body: JSON.stringify(req)})
  }
}