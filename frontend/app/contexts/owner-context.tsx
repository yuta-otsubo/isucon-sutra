import { useNavigate, useSearchParams } from "@remix-run/react";
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  OwnerGetChairsResponse,
  OwnerGetSalesResponse,
  fetchOwnerGetChairs,
  fetchOwnerGetSales,
} from "~/apiClient/apiComponents";
import { getCookieValue } from "~/components/modules/session/cookie";
import { isClientApiError } from "~/types";

type ClientProviderRequest = Partial<{
  chairs: OwnerGetChairsResponse["chairs"];
  sales: OwnerGetSalesResponse;
  provider?: {
    id: string;
    name: string;
  };
}>;

const DUMMY_DATA = {
  total_sales: 8087,
  chairs: [
    { id: "chair-a", name: "椅子A", sales: 999 },
    { id: "chair-b", name: "椅子B", sales: 999 },
  ],
  models: [
    { model: "モデルA", sales: 999 },
    { model: "モデルB", sales: 999 },
  ],
} as const satisfies OwnerGetSalesResponse;

const ClientProviderContext = createContext<Partial<ClientProviderRequest>>({});

const timestamp = (date: string) => Math.floor(new Date(date).getTime() / 1000);

export const ProviderProvider = ({ children }: { children: ReactNode }) => {
  // TODO:
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const id = searchParams.get("id") ?? undefined;
  const name = searchParams.get("name") ?? undefined;
  const since = searchParams.get("since") ?? undefined;
  const until = searchParams.get("until") ?? undefined;

  useEffect(() => {
    if (getCookieValue(document.cookie, "app_session") === undefined) {
      navigate("/client/register");
    }
  }, [navigate]);

  const isDummy = useMemo(() => {
    try {
      const isDummy = sessionStorage.getItem("is-dummy-for-provider");
      return isDummy === "true";
    } catch (e) {
      if (typeof e === "string") {
        console.error(`CONSOLE ERROR: ${e}`);
      }
      return false;
    }
  }, []);

  const [chairs, setChairs] = useState<OwnerGetChairsResponse>();
  const [sales, setSales] = useState<OwnerGetSalesResponse>();

  useEffect(() => {
    if (isDummy) {
      setSales({
        total_sales: DUMMY_DATA.total_sales,
        chairs: DUMMY_DATA.chairs,
        models: DUMMY_DATA.models,
      });
    } else {
      const abortController = new AbortController();
      Promise.all([
        fetchOwnerGetChairs({}, abortController.signal).then((res) =>
          setChairs(res),
        ),
        since && until
          ? fetchOwnerGetSales(
              {
                queryParams: {
                  since: timestamp(since),
                  until: timestamp(until),
                },
              },
              abortController.signal,
            ).then((res) => setSales(res))
          : Promise.resolve(),
      ]).catch((e) => {
        console.error(`ERROR: ${JSON.stringify(e)}`);
        if (isClientApiError(e)) {
          if (e.stack.status === 401) {
            navigate("/owner/register");
          }
        }
      });
      return () => {
        abortController.abort();
      };
    }
  }, [setChairs, setSales, since, until, isDummy, navigate]);

  const responseClientProvider = useMemo<ClientProviderRequest>(() => {
    return {
      chairs: chairs?.chairs ?? [],
      sales,
      provider: id && name ? { id, name } : undefined,
    };
  }, [chairs, sales, id, name]);

  return (
    <ClientProviderContext.Provider value={responseClientProvider}>
      {children}
    </ClientProviderContext.Provider>
  );
};

export const useClientProviderContext = () => useContext(ClientProviderContext);
