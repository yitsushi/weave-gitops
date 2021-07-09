import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  children?: any;
} & React.HTMLProps<HTMLFormElement>;

function Form({ className, children, ...rest }: Props) {
  return (
    <form {...rest} className={className}>
      {children}
    </form>
  );
}

export default styled(Form)``;
