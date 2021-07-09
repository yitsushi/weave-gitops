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
    name: "",
    namespace: "",
    url: "",
    path: "",
    branch: "",
    deploymentType: AddApplicationRequestDeploymentType.KUSTOMIZE,
    chart: "",
    sourceType: AddApplicationRequestSourceType.GIT,
    appConfigUrl: "",
    dryRun: false,
    autoMerge: false,
  });

  const submit = () => {
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
        <Button type="submit">
          {loading ? <CircularProgress /> : "Submit"}
        </Button>
      </Form>
    </Page>
  );
}

export default styled(ApplicationAdd)``;
