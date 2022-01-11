import * as React from "react";
import { Link as RouterLink } from "react-router-dom";
import styled from "styled-components";
import Text from "./Text";

type Props = {
  className?: string;
  to?: string;
  innerRef?: any;
  children?: any;
  href?: any;
  onClick?: (ev: any) => void;
};

function Link({
  children,
  href,
  onClick,
  className,
  to = "",
  ...props
}: Props) {
  const txt = <Text color="primary">{children}</Text>;

  if (href) {
    return (
      <a onClick={onClick} className={className} href={href}>
        {txt}
      </a>
    );
  }

  return (
    <RouterLink onClick={onClick} className={className} to={to} {...props}>
      {txt}
    </RouterLink>
  );
}

export default styled(Link)`
  text-decoration: none;
  &.title-bar-button {
    width: 250px;
  }
`;
