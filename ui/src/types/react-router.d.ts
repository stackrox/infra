/* eslint no-var: 0, vars-on-top: 0, no-redeclare: 0, import/export: 0, @typescript-eslint/no-explicit-any: 0 */

/**
 * These definitions are copied from the build artifacts of react-router branch
 * https://github.com/ReactTraining/react-router/tree/ts.
 *
 * TODO:
 *  - remove this file as soon as react-router@6 is released
 *  - also `yarn remove @types/react-router
 */

import * as React from 'react';
import PropTypes from 'prop-types';
import { Path, State, PathPieces, Location, Blocker, To, History, InitialEntry } from 'history';

declare module 'react-router' {
  /**
   * A <Router> that stores all entries in memory.
   */
  export function MemoryRouter({
    children,
    initialEntries,
    initialIndex,
    timeout,
  }: MemoryRouterProps): React.ReactElement;
  export namespace MemoryRouter {
    var displayName: string;
    var propTypes: {
      children: PropTypes.Requireable<PropTypes.ReactNodeLike>;
      timeout: PropTypes.Requireable<number>;
      initialEntries: PropTypes.Requireable<
        (
          | string
          | PropTypes.InferProps<{
              pathname: PropTypes.Requireable<string>;
              search: PropTypes.Requireable<string>;
              hash: PropTypes.Requireable<string>;
              state: PropTypes.Requireable<object>;
              key: PropTypes.Requireable<string>;
            }>
          | null
          | undefined
        )[]
      >;
      initialIndex: PropTypes.Requireable<number>;
    };
  }
  export interface MemoryRouterProps {
    children?: React.ReactNode;
    initialEntries?: InitialEntry[];
    initialIndex?: number;
    timeout?: number;
  }
  /**
   * Navigate programmatically using a component.
   */
  export function Navigate({ to, replace, state }: NavigateProps): null;
  export namespace Navigate {
    var displayName: string;
    var propTypes: {
      to: PropTypes.Validator<
        | string
        | PropTypes.InferProps<{
            pathname: PropTypes.Requireable<string>;
            search: PropTypes.Requireable<string>;
            hash: PropTypes.Requireable<string>;
          }>
      >;
      replace: PropTypes.Requireable<boolean>;
      state: PropTypes.Requireable<object>;
    };
  }
  export interface NavigateProps {
    to: To;
    replace?: boolean;
    state?: State;
  }
  /**
   * Renders the child route's element, if there is one.
   */
  export function Outlet(): React.ReactElement | null;
  export namespace Outlet {
    var displayName: string;
    var propTypes: {};
  }
  /**
   * Used in a route config to render an element.
   */
  export function Route({ element }: RouteProps): React.ReactElement | null;
  export namespace Route {
    var displayName: string;
    var propTypes: {
      children: PropTypes.Requireable<PropTypes.ReactNodeLike>;
      element: PropTypes.Requireable<PropTypes.ReactElementLike>;
      path: PropTypes.Requireable<string>;
    };
  }
  export interface RouteProps {
    children?: React.ReactNode;
    element?: React.ReactElement | null;
    path?: string;
  }
  /**
   * The root context provider. There should be only one of these in a given app.
   */
  export function Router({
    children,
    history,
    static: staticProp,
    timeout,
  }: RouterProps): React.ReactElement;
  export namespace Router {
    var displayName: string;
    var propTypes: {
      children: PropTypes.Requireable<PropTypes.ReactNodeLike>;
      history: PropTypes.Requireable<
        PropTypes.InferProps<{
          action: PropTypes.Requireable<string>;
          location: PropTypes.Requireable<object>;
          push: PropTypes.Requireable<(...args: any[]) => any>;
          replace: PropTypes.Requireable<(...args: any[]) => any>;
          go: PropTypes.Requireable<(...args: any[]) => any>;
          listen: PropTypes.Requireable<(...args: any[]) => any>;
          block: PropTypes.Requireable<(...args: any[]) => any>;
        }>
      >;
      timeout: PropTypes.Requireable<number>;
    };
  }
  export interface RouterProps {
    children?: React.ReactNode;
    history: History;
    static?: boolean;
    timeout?: number;
  }
  /**
   * A wrapper for useRoutes that treats its children as route and/or redirect
   * objects.
   */
  export function Routes({
    basename,
    caseSensitive,
    children,
  }: RoutesProps): React.ReactElement | null;
  export namespace Routes {
    var displayName: string;
    var propTypes: {
      basename: PropTypes.Requireable<string>;
      caseSensitive: PropTypes.Requireable<boolean>;
      children: PropTypes.Requireable<PropTypes.ReactNodeLike>;
    };
  }
  export interface RoutesProps {
    basename?: string;
    caseSensitive?: boolean;
    children?: React.ReactNode;
  }
  /**
   * Blocks all navigation attempts. This is useful for preventing the page from
   * changing until some condition is met, like saving form data.
   */
  export function useBlocker(blocker: Blocker, when?: boolean): void;
  /**
   * Returns the full href for the given "to" value. This is useful for building
   * custom links that are also accessible and preserve right-click behavior.
   */
  export function useHref(to: To): string;
  /**
   * Returns true if this component is a descendant of a <Router>.
   */
  export function useInRouterContext(): boolean;
  /**
   * Returns the current location object, which represents the current URL in web
   * browsers.
   *
   * NOTE: If you're using this it may mean you're doing some of your own
   * "routing" in your app, and we'd like to know what your use case is. We may be
   * able to provide something higher-level to better suit your needs.
   */
  export function useLocation(): Location;
  /**
   * Returns true if the router is pending a location update.
   */
  export function useLocationPending(): boolean;
  /**
   * Returns true if the URL for the given "to" value matches the current URL.
   * This is useful for components that need to know "active" state, e.g.
   * <NavLink>.
   */
  export function useMatch(to: To): boolean;
  /**
   * The interface for the navigate() function returned from useNavigate().
   */
  export interface NavigateFunction {
    (
      to: To | number,
      options?: {
        replace?: boolean;
        state?: State | null;
      }
    ): void;
  }
  /**
   * Returns an imperative method for changing the location. Used by <Link>s, but
   * may also be used by other elements to change the location.
   */
  export function useNavigate(): NavigateFunction;
  /**
   * Returns the outlet element at this level of the route hierarchy. Used to
   * render child routes.
   */
  export function useOutlet(): React.ReactElement | null;
  /**
   * Returns a hash of the dynamic params that were matched in the route path.
   * This is useful for using ids embedded in the URL to fetch data, but we
   * eventually want to provide something at a higher level for this.
   */
  export function useParams(): Params;
  /**
   * Returns a fully-resolved location object relative to the current location.
   */
  export function useResolvedLocation(to: To): ResolvedLocation;
  /**
   * Returns the element of the route that matched the current location, prepared
   * with the correct context to render the remainder of the route tree. Route
   * elements in the tree must render an <Outlet> to render their child route's
   * element.
   */
  export function useRoutes(
    routes: PartialRouteObject[],
    basename?: string,
    caseSensitive?: boolean
  ): React.ReactElement | null;
  /**
   * Utility function that creates a routes config object from an array of
   * PartialRouteObject objects.
   */
  export function createRoutesFromArray(array: PartialRouteObject[]): RouteObject[];
  /**
   * Utility function that creates a routes config object from a React "children"
   * object, which is usually either a <Route> element or an array of them.
   */
  export function createRoutesFromChildren(children: React.ReactNode): RouteObject[];
  /**
   * A "partial route" object is usually supplied by the user and may omit certain
   * properties of a real route object such as `path` and `element`, which have
   * reasonable defaults.
   */
  export interface PartialRouteObject {
    path?: string;
    element?: React.ReactNode;
    children?: PartialRouteObject[];
  }
  /**
   * A route object represents a logical route, with (optionally) its child routes
   * organized in a tree-like structure.
   */
  export interface RouteObject {
    path: string;
    element: React.ReactNode;
    children?: RouteObject[];
  }
  /**
   * Creates a path with params interpolated.
   */
  export function generatePath(pathname: string, params?: Params): string;
  /**
   * The parameters that were parsed from the URL path.
   */
  export type Params = Record<string, string>;
  /**
   * Matches the given routes to a location and returns the match data.
   */
  export function matchRoutes(
    routes: PartialRouteObject[],
    location: Path | PathPieces,
    basename?: string,
    caseSensitive?: boolean
  ): MatchObject[] | null;
  export interface MatchObject {
    params: Params;
    pathname: string;
    route: RouteObject;
  }
  /**
   * Returns a fully resolved location object relative to the given pathname.
   */
  export function resolveLocation(to: To, fromPathname?: string): ResolvedLocation;
  export type ResolvedLocation = Omit<Location, 'state' | 'key'>;
}
