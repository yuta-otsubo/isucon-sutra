import { ReactNode, createContext, useContext } from "react";

export type AccessToken = string;

/**
 * フロント側で利用するクライアント情報
 */
type ClientInfo = {
  id: string;
  name: string;
  accessToken: string;
};

const clientContext = createContext<ClientInfo>({
  id: "",
  name: "",
  accessToken: "",
});

export const ClientProvider = ({
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
    <clientContext.Provider value={fetchedValue}>
      {children}
    </clientContext.Provider>
  );
};

export const useClient = () => useContext(clientContext);
