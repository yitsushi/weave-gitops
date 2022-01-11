/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../applications/fetch.pb"

export enum SourceRefKind {
  GitRepository = "GitRepository",
  Bucket = "Bucket",
  HelmRepository = "HelmRepository",
}

export type Interval = {
  hours?: string
  minutes?: string
  seconds?: string
}

export type SourceRef = {
  kind?: SourceRefKind
  name?: string
}

export type Kustomization = {
  name?: string
  namespace?: string
  path?: string
  sourceRef?: SourceRef
  interval?: Interval
}

export type AddKustomizationRequest = {
  repoName?: string
  appName?: string
  name?: string
  namespace?: string
  path?: string
  sourceRef?: SourceRef
  interval?: Interval
}

export type AddKustomizationResponse = {
  success?: boolean
  kustomization?: Kustomization
}

export type RemoveKustomizationRequest = {
  repoName?: string
  appName?: string
  kustomizationName?: string
}

export type RemoveKustomizationResponse = {
  success?: boolean
}

export class AppKustomization {
  static Add(req: AddKustomizationRequest, initReq?: fm.InitReq): Promise<AddKustomizationResponse> {
    return fm.fetchReq<AddKustomizationRequest, AddKustomizationResponse>(`/v1/repo/${req["repoName"]}/app/${req["appName"]}/kustomization`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static Remove(req: RemoveKustomizationRequest, initReq?: fm.InitReq): Promise<RemoveKustomizationResponse> {
    return fm.fetchReq<RemoveKustomizationRequest, RemoveKustomizationResponse>(`/v1/repo/${req["repoName"]}/app/${req["appName"]}/kustomization/${req["kustomizationName"]}`, {...initReq, method: "DELETE"})
  }
}