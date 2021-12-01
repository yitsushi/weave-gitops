import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
} from "../lib/api/applications/applications.pb";
import { GrpcErrorCodes } from "../lib/types";

function poller(cb, interval) {
  if (process.env.NODE_ENV === "test") {
    // Stay synchronous in tests
    return cb();
  }

  return setInterval(cb, interval);
}

function superPoll(cb, interval) {
  let i = interval;
  let timeout = null;

  const poll = () => {
    cb();
    timeout = setTimeout(() => {
      poll();
    }, i);
  };

  poll();

  return {
    update: (newInterval) => {
      i = newInterval;
    },
    cancel: () => {
      clearTimeout(timeout);
    },
    debug: () => console.log(i),
  };
}

export function myfunc(p: Promise<any>) {
  let i = 1000;
  const { update, cancel, debug } = superPoll(
    () =>
      p
        .then(() => {
          console.log("continuing polling");
          // cancel();
          debug();
          i = i + 1000;
          update(i);
        })
        .catch((e) => {
          // update(e.newInterval);
        }),
    1000
  );
}

export function useIsAuthenticated(provider: GitProvider): boolean {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const { getProviderToken } = useContext(AppContext);

  useEffect(() => {
    const token = getProviderToken(provider);

    if (token) {
      setIsAuthenticated(true);
    }
  }, [provider]);

  return isAuthenticated;
}

export default function useAuth() {
  const [loading, setLoading] = useState(true);
  const { applicationsClient, getProviderToken, storeProviderToken } =
    useContext(AppContext);

  const getGithubDeviceCode = () => {
    setLoading(true);
    return applicationsClient
      .GetGithubDeviceCode({})
      .finally(() => setLoading(false));
  };

  const getGithubAuthStatus = (codeRes: GetGithubDeviceCodeResponse) => {
    let poll;
    return {
      cancel: () => clearInterval(poll),
      promise: new Promise<GetGithubAuthStatusResponse>((accept, reject) => {
        poll = poller(() => {
          applicationsClient
            .GetGithubAuthStatus(codeRes)
            .then((res) => {
              clearInterval(poll);
              accept(res);
            })
            .catch(({ code, message }) => {
              // Unauthenticated means we can keep polling.
              //  On anything else, stop polling and report.
              if (code !== GrpcErrorCodes.Unauthenticated) {
                clearInterval(poll);
                reject({ message });
              }
            });
        }, (codeRes.interval + 1) * 100);
      }),
    };
  };

  return {
    loading,
    getGithubDeviceCode,
    getGithubAuthStatus,
    getProviderToken,
    storeProviderToken,
  };
}
