import { useSearchParams } from "@remix-run/react";
import { createContext, useContext, useMemo, type ReactNode } from "react";
import {
  useAppGetNotification,
  type AppGetNotificationError,
} from "~/apiClient/apiComponents";
import type {
  AppRequest,
  Coordinate,
  RequestStatus,
} from "~/apiClient/apiSchemas";
import type { User } from "~/types";

const UserContext = createContext<Partial<User>>({});

const RequestContext = createContext<{
  data:
    | AppRequest
    | { status?: RequestStatus; destination_coordinate?: Coordinate };
  error?: AppGetNotificationError | null;
  isLoading: boolean;
}>({ isLoading: false, data: { status: undefined } });

const RequestProvider = ({
  children,
  accessToken,
}: {
  children: ReactNode;
  accessToken: string;
}) => {
  const notificationResponse = useAppGetNotification({
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "text/event-stream",
    },
  });
  const { data, error, isLoading } = notificationResponse;
  // react-queryでstatusCodeが取れない && 現状statusCode:204はBlobで帰ってくる
  const [searchParams] = useSearchParams();
  const responseData = useMemo(() => {
    const status = (searchParams.get("debug_status") ?? undefined) as
      | RequestStatus
      | undefined;
    const destination_coordinate = ((): Coordinate | undefined => {
      // expected format: 123,456
      const v = searchParams.get("debug_destination_coordinate") ?? "";
      const m = v.match(/(\d+),(\d+)/);
      if (!m) return;
      return { latitude: Number(m[1]), longitude: Number(m[2]) };
    })();

    let fetchedData: Partial<AppRequest> = data ?? {};
    if (data instanceof Blob) {
      fetchedData = {};
    }

    // TODO:
    return { ...fetchedData, status, destination_coordinate } as AppRequest;
  }, [data, searchParams]);

  /**
   * TODO: SSE処理
   */

  return (
    <RequestContext.Provider value={{ data: responseData, error, isLoading }}>
      {children}
    </RequestContext.Provider>
  );
};

export const UserProvider = ({ children }: { children: ReactNode }) => {
  // TODO:
  const [searchParams] = useSearchParams();
  const accessTokenParameter = searchParams.get("access_token");
  const userIdParameter = searchParams.get("id");

  const user: Partial<User> = useMemo(() => {
    if (accessTokenParameter !== null && userIdParameter !== null) {
      requestIdleCallback(() => {
        sessionStorage.setItem("user_access_token", accessTokenParameter);
        sessionStorage.setItem("user_id", userIdParameter);
      });
      return {
        accessToken: accessTokenParameter,
        id: userIdParameter,
        name: "ISUCON太郎",
      };
    }
    const accessToken =
      sessionStorage.getItem("user_access_token") ?? undefined;
    const id = sessionStorage.getItem("user_id") ?? undefined;
    return {
      accessToken,
      id,
      name: "ISUCON太郎",
    };
  }, [accessTokenParameter, userIdParameter]);

  return (
    <UserContext.Provider value={user}>
      <RequestProvider accessToken={user.accessToken ?? ""}>
        {children}
      </RequestProvider>
    </UserContext.Provider>
  );
};

export const useUser = () => useContext(UserContext);

export const useRequest = () => useContext(RequestContext);
