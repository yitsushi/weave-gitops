import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Button from "../components/Button";
import Form from "../components/Form";
import Input from "../components/Input";
import Page from "../components/Page";
import useApplications from "../hooks/applications";
import {
  AddApplicationRequest,
  AddApplicationRequestDeploymentType,
  AddApplicationRequestSourceType,
} from "../lib/api/applications/applications.pb";

type Props = {
  className?: string;
};

function ApplicationAdd({ className }: Props) {
  const { addApplication, loading } = useApplications();
  const [formState, setFormState] = React.useState<AddApplicationRequest>({
    name: "my-app",
    namespace: "wego-system",
    url: "ssh://git@github.com/jpellizzari/stringly",
    path: "k8s/base/apps",
    branch: "main",
    deploymentType: AddApplicationRequestDeploymentType.kustomize,
    chart: "",
    sourceType: AddApplicationRequestSourceType.git,
    appConfigUrl: "",
    dryRun: false,
    autoMerge: false,
  });

  const submit = () => {
    console.log(formState);
    addApplication(formState);
  };

  return (
    <Page title="Add Application" className={className}>
      <Form
        onSubmit={(ev) => {
          ev.preventDefault();
          submit();
        }}
      >
        <Input
          onChange={(ev: React.FormEvent<HTMLInputElement>) => {
            setFormState({
              ...formState,
              name: ev.currentTarget.value,
            });
          }}
          name="name"
          value={formState.name}
        />
        <Input
          onChange={(ev: React.FormEvent<HTMLInputElement>) => {
            setFormState({
              ...formState,
              namespace: ev.currentTarget.value,
            });
          }}
          name="name"
          value={formState.namespace}
        />
        <Input
          onChange={(ev: React.FormEvent<HTMLInputElement>) => {
            setFormState({
              ...formState,
              url: ev.currentTarget.value,
            });
          }}
          name="name"
          value={formState.url}
        />
        <Button type="submit">
          {loading ? <CircularProgress /> : "Submit"}
        </Button>
      </Form>
    </Page>
  );
}

export default styled(ApplicationAdd)``;
