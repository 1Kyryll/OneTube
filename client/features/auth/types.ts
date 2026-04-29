export type User = {
  id: string;
  email: string;
  username: string;
  display_name: string;
  avatar?: string;
};

export type SignupInput = {
  email: string;
  username: string;
  display_name: string;
  password: string;
};

export type LoginInput = {
  email: string;
  password: string;
};
