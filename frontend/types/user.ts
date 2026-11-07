export type UserRole = "user" | "admin";

export type User = {
  id: number;
  email: string;
  display_name: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
};

export type UserResponse = {
  data: User;
};
