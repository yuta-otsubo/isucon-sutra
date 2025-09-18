import { useEmulator } from "~/components/hooks/emulate";
import { List } from "~/components/modules/list/list";
import { ListItem } from "~/components/modules/list/list-item";
import { useSimulatorContext } from "~/contexts/simulator-context";
import { ChairInfo } from "./ChairInfo";

export default function Index() {
  const { targetChair } = useSimulatorContext();
  useEmulator(targetChair);
  return (
    <div className="p-6">
      {targetChair !== undefined ? (
        <List>
          <ListItem key={targetChair.id}>
            <ChairInfo chair={targetChair} />
          </ListItem>
        </List>
      ) : null}
    </div>
  );
}
