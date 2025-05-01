import { ReactNode, createContext, useContext } from "react";

export type AccessToken = string;

/**
 * フロント側で利用するクライアント情報
 */
type UserInfo = {
  id: string;
  name: string;
  accessToken: string;
};

const userContext = createContext<UserInfo>({
  id: "",
  name: "",
  accessToken: "",
});

export const UserProvider = ({
  children,
  accessToken,
}: {
  children: ReactNode;
  accessToken: string;
}) => {
  /**
   * openapi上にfetchするものがないので一旦仮置き
   */
  const fetchedValue = { id: "fetched-id", name: "fetched-name", accessToken };

  return (
    <userContext.Provider value={fetchedValue}>
      {children}
    </userContext.Provider>
  );
};

export const useClient = () => useContext(userContext);
