import MaterialInput from "@material-ui/core/Input";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  label?: string;
  name: string;
} & React.HTMLProps<HTMLInputElement>;

function Input({ className, name, label, ...rest }: Props) {
  return (
    <div className={className}>
      {label && <label htmlFor={name}>{label}</label>}
      <MaterialInput inputProps={rest} color="primary" />
    </div>
  );
}

export default styled(Input)``;
