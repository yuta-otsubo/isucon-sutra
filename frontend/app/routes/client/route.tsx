import { Outlet } from "@remix-run/react";
import { CircleIcon } from "~/components/icon/circle";
import { FooterNavigation } from "~/components/modules/footer-navigation/footer-navigation";
import { UserProvider } from "../../contexts/user-context";

export default function ClientLayout() {
  return (
    <UserProvider>
      <Outlet />
      <FooterNavigation
        navigationMenus={[
          { icon: CircleIcon, link: "/client", label: "ride" },
          { icon: CircleIcon, link: "/client/history", label: "history" },
        ]}
      />
    </UserProvider>
  );
}
