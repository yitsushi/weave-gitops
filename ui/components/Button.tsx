import MaterialButton, { ButtonProps } from "@material-ui/core/Button";
import * as React from "react";
import styled from "styled-components";

function Button({ className, ...rest }: ButtonProps) {
  return <MaterialButton {...rest} className={className}></MaterialButton>;
}

export default styled(Button)``;
