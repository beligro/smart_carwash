import styled from 'styled-components';

// Container components
export const Container = styled.div`
  max-width: 100%;
  padding: 16px;
  margin: 0 auto;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
    Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
`;

export const Card = styled.div`
  background-color: var(--tg-theme-secondary-bg-color, #f5f5f5);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
`;

export const Header = styled.header`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
`;

export const Footer = styled.footer`
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: var(--tg-theme-hint-color, #999999);
`;

// Typography components
export const Title = styled.h1`
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: var(--tg-theme-text-color, #000000);
`;

export const Subtitle = styled.h2`
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: var(--tg-theme-text-color, #000000);
`;

export const Text = styled.p`
  font-size: 16px;
  line-height: 1.5;
  margin: 0 0 16px 0;
  color: var(--tg-theme-text-color, #000000);
`;

export const SmallText = styled.p`
  font-size: 14px;
  line-height: 1.4;
  margin: 0 0 8px 0;
  color: var(--tg-theme-hint-color, #999999);
`;

// Button components
export const Button = styled.button`
  background-color: var(--tg-theme-button-color, #2481cc);
  color: var(--tg-theme-button-text-color, #ffffff);
  border: none;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s ease;
  
  &:hover {
    opacity: 0.9;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

export const SecondaryButton = styled(Button)`
  background-color: transparent;
  color: var(--tg-theme-button-color, #2481cc);
  border: 1px solid var(--tg-theme-button-color, #2481cc);
`;

// Layout components
export const Flex = styled.div`
  display: flex;
  align-items: center;
`;

export const FlexColumn = styled(Flex)`
  flex-direction: column;
  align-items: flex-start;
`;

export const FlexBetween = styled(Flex)`
  justify-content: space-between;
`;

export const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
`;

// Box status components
export const BoxItem = styled.div<{ status: string }>`
  background-color: ${props => {
    switch (props.status) {
      case 'available':
        return '#4CAF50';
      case 'occupied':
        return '#F44336';
      case 'maintenance':
        return '#FFC107';
      default:
        return '#E0E0E0';
    }
  }};
  color: white;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
  font-weight: 600;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

export const StatusIndicator = styled.span<{ status: string }>`
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-right: 8px;
  background-color: ${props => {
    switch (props.status) {
      case 'available':
        return '#4CAF50';
      case 'occupied':
        return '#F44336';
      case 'maintenance':
        return '#FFC107';
      default:
        return '#E0E0E0';
    }
  }};
`;

// Loading component
export const LoadingSpinner = styled.div`
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top: 4px solid var(--tg-theme-button-color, #2481cc);
  width: 24px;
  height: 24px;
  animation: spin 1s linear infinite;
  margin: 16px auto;
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;
