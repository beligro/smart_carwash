import React from 'react';
import styled from 'styled-components';

const HeaderContainer = styled.header`
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: background-color 0.3s ease, color 0.3s ease;
`;

const Logo = styled.div`
  font-size: 24px;
  font-weight: bold;
  display: flex;
  align-items: center;
`;

const LogoIcon = styled.span`
  margin-right: 8px;
  font-size: 28px;
`;

const Header = ({ theme }) => {
  return (
    <HeaderContainer theme={theme}>
      <Logo>
        <LogoIcon>🚿</LogoIcon>
        Умная автомойка
      </Logo>
    </HeaderContainer>
  );
};

export default Header;
