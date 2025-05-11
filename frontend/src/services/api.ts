import axios from 'axios';

// Create axios instance with base URL
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8000/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Types
export interface User {
  id: number;
  telegram_id: string;
  username: string | null;
  first_name: string | null;
  last_name: string | null;
  role: string;
  created_at: string;
  updated_at: string | null;
}

export interface CarwashBox {
  id: number;
  box_number: number;
  status: string;
  created_at: string;
  updated_at: string | null;
}

export interface CarwashInfo {
  total_boxes: number;
  available_boxes: number;
  available_box_numbers: number[];
}

// API functions
export const createUser = async (userData: {
  telegram_id: string;
  username?: string;
  first_name?: string;
  last_name?: string;
}): Promise<User> => {
  const response = await api.post<User>('/users', userData);
  return response.data;
};

export const getCarwashInfo = async (): Promise<CarwashInfo> => {
  const response = await api.get<CarwashInfo>('/carwash/info');
  return response.data;
};

export const getAllBoxes = async (): Promise<CarwashBox[]> => {
  const response = await api.get<CarwashBox[]>('/carwash/boxes');
  return response.data;
};

export default {
  createUser,
  getCarwashInfo,
  getAllBoxes,
};
