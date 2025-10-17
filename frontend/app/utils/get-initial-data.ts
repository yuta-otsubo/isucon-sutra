type InitialChair = {
  id: string;
  owner_id: string;
  name: string;
  model: string;
  token: string;
};

type InitialOwner = {
  id: string;
  name: string;
  token: string;
};

const initialOwnerData = __INITIAL_OWNER_DATA__;

export const getOwners = (): InitialOwner[] => {
  return (
    initialOwnerData?.owners?.map((owner) => ({
      ...owner,
    })) ?? []
  );
};

export const getSimulateChair = (): InitialChair | undefined => {
  return initialOwnerData?.targetSimulatorChair;
};
