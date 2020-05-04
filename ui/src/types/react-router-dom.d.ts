/* eslint no-var: 0, vars-on-top: 0, no-redeclare: 0, import/export: 0, @typescript-eslint/no-explicit-any: 0 */

/**
 * These definitions are copied from the build artifacts of react-router branch
 * https://github.com/ReactTraining/react-router/tree/ts.
 *
 * TODO:
 *  - remove this file as soon as react-router@6 is released
 *  - also `yarn remove @types/react-router-dom
 */

import * as React from 'react';
import PropTypes from 'prop-types';
import { State, To } from 'history';
import {
  MemoryRouter,
  Navigate,
  Outlet,
  Route,
  Router,
  Routes,
  useBlocker,
  useHref,
  useInRouterContext,
  useLocation,
  useLocationPending,
  useMatch,
  useNavigate,
  useOutlet,
  useParams,
  useResolvedLocation,
  useRoutes,
  createRoutesFromArray,
  createRoutesFromChildren,
  generatePath,
  matchRoutes,
  resolveLocation,
} from 'react-router';

declare module 'react-router-dom' {
  export {
    MemoryRouter,
    Navigate,
    Outlet,
    Route,
    Router,
    Routes,
    useBlocker,
    useHref,
    useInRouterContext,
    useLocation,
    useLocationPending,
    useMatch,
    useNavigate,
    useOutlet,
    useParams,
    useResolvedLocation,
    useRoutes,
    createRoutesFromArray,
    createRoutesFromChildren,
    generatePath,
    matchRoutes,
    resolveLocation,
  };
  /**
   * A <Router> for use in web browsers. Provides the cleanest URLs.
   */
  export function BrowserRouter({ children, timeout, window }: BrowserRouterProps): JSX.Element;
  export namespace BrowserRouter {
    var displayName: string;
    var propTypes: {
      children: PropTypes.Requireable<PropTypes.ReactNodeLike>;
      timeout: PropTypes.Requireable<number>;
      window: PropTypes.Requireable<object>;
    };
  }
  export interface BrowserRouterProps {
    children?: React.ReactNode;
    timeout?: number;
    window?: Window;
  }
  /**
   * A <Router> for use in web browsers. Stores the location in the hash
   * portion of the URL so it is not sent to the server.
   */
  export function HashRouter({ children, timeout, window }: HashRouterProps): JSX.Element;
  export namespace HashRouter {
    var displayName: string;
    var propTypes: {
      children: PropTypes.Requireable<PropTypes.ReactNodeLike>;
      timeout: PropTypes.Requireable<number>;
      window: PropTypes.Requireable<object>;
    };
  }
  export interface HashRouterProps {
    children?: React.ReactNode;
    timeout?: number;
    window?: Window;
  }
  /**
   * The public API for rendering a history-aware <a>.
   */
  export const Link: React.ForwardRefExoticComponent<
    LinkProps & React.RefAttributes<HTMLAnchorElement>
  >;
  export interface LinkProps extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> {
    replace?: boolean;
    state?: State;
    to: To;
  }
  /**
   * A <Link> wrapper that knows if it's "active" or not.
   */
  export const NavLink: React.ForwardRefExoticComponent<
    NavLinkProps & React.RefAttributes<HTMLAnchorElement>
  >;
  export interface NavLinkProps extends LinkProps {
    activeClassName?: string;
    activeStyle?: object;
  }
  /**
   * A declarative interface for showing a window.confirm dialog with the given
   * message when the user tries to navigate away from the current page.
   *
   * This also serves as a reference implementation for anyone who wants to
   * create their own custom prompt component.
   */
  export function Prompt({ message, when }: PromptProps): null;
  export namespace Prompt {
    var displayName: string;
    var propTypes: {
      message: PropTypes.Requireable<string>;
      when: PropTypes.Requireable<boolean>;
    };
  }
  export interface PromptProps {
    message: string;
    when?: boolean;
  }
  /**
   * Prevents navigation away from the current page using a window.confirm prompt
   * with the given message.
   */
  export function usePrompt(message: string, when?: boolean): void;
  /**
   * A convenient wrapper for reading and writing search parameters via the
   * URLSearchParams interface.
   */
  export function useSearchParams(
    defaultInit: URLSearchParamsInit
  ): (URLSearchParams | ((nextInit: any, navigateOpts: any) => void))[];
  /**
   * Creates a URLSearchParams object using the given initializer.
   *
   * This is identical to `new URLSearchParams(init)` except it also
   * supports arrays as values in the object form of the initializer
   * instead of just strings. This is convenient when you need multiple
   * values for a given key, but don't want to use an array initializer.
   *
   * For example, instead of:
   *
   *   let searchParams = new URLSearchParams([
   *     ['sort', 'name'],
   *     ['sort', 'price']
   *   ]);
   *
   * you can do:
   *
   *   let searchParams = createSearchParams({
   *     sort: ['name', 'price']
   *   });
   */
  export function createSearchParams(init?: URLSearchParamsInit): URLSearchParams;
  export type ParamKeyValuePair = [string, string];
  export type URLSearchParamsInit =
    | string
    | ParamKeyValuePair[]
    | Record<string, string | string[]>
    | URLSearchParams;
}
