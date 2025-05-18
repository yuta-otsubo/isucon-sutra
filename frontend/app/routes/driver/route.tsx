import type { MetaFunction } from "@remix-run/node";
import { Outlet } from "@remix-run/react";
import { FooterNavigation } from "~/components/FooterNavigation";
import { CircleIcon } from "~/components/icon/circle";
import { DriverProvider } from "../../contexts/driver-context";

export const meta: MetaFunction = () => {
  return [
    { title: "椅子 | ISURIDE" },
    { name: "description", content: "isucon14" },
  ];
};

export default function DriverLayout() {
  return (
    <>
      <div className="font-sans p-4">
        <DriverProvider>
          <Outlet />
        </DriverProvider>
      </div>
      <FooterNavigation
        navigationMenus={[
          { icon: CircleIcon, link: "/driver/", label: "ride" },
          { icon: CircleIcon, link: "/driver/history", label: "history" },
        ]}
      />
    </>
  );
}
