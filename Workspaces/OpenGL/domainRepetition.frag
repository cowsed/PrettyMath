#version 330
in vec2 UV;
layout(location = 0) out vec4 frag_color;

//Uniforms
uniform vec3 clearColor;
uniform vec3 camPos;
uniform vec3 lookAt;


uniform float shadowK;
uniform float shadowEpsilon;
//Rendering things
uniform int MAX_MARCHING_STEPS;
uniform float MAX_DIST = 100.0;
//Control over scene
uniform int DBDRAW;

uniform vec3 lightPos;
uniform vec3 boxPos;

uniform vec3 repeat;
uniform vec3 numRepeats;
//Constants
const float MIN_DIST = 0.0;
const float EPSILON = 0.0001;

uniform float ambient;
const float gamma = 2.2;


struct sdResult{
    float distance;
    int materialID;
};

sdResult sdRound(sdResult sd, float radius){
    sd.distance-=radius;
    return sd;
}

sdResult sdSmoothUnion(sdResult a, sdResult b, float k){
    float d1 = a.distance;
    float d2 = b.distance;
    float h = clamp( 0.5 + 0.5*(d2-d1)/k, 0.0, 1.0 );
    sdResult res;
    res.distance = mix( d2, d1, h ) - k*h*(1.0-h);
    res.materialID = h>.5 ? a.materialID : b.materialID;
    return res;
}
sdResult sdMin(sdResult a, sdResult b){
    if (a.distance<b.distance){
        return a;
    }
    return b;
}
sdResult sdMax(sdResult a, sdResult b){
    if (a.distance>b.distance){
        return a;
    }
    return b;
}

sdResult sdSub(sdResult a, sdResult b){
    a.distance = -a.distance;
    if (a.distance>b.distance){
        return a;
    }
    return b;
}

sdResult sdSphere(vec3 samplePoint, float radius, int matID) {
    sdResult res;
    res.distance = length(samplePoint) - radius;
    res.materialID = matID;
    return res;
}

sdResult sdBox( vec3 p, vec3 b, int matID)
{
  vec3 q = abs(p) - b;
  sdResult res;
  res.distance = length(max(q,0.0)) + min(max(q.x,max(q.y,q.z)),0.0);
  res.materialID = matID;
  return res;
}

sdResult sdXYPlane(vec3 p, float z, int matID){
    sdResult res;
    res.distance = abs(p.z-z);
    res.materialID = matID;
    return res;
}

vec3 opRepLim(vec3 p,float s, in vec3 lima, in vec3 limb )
{
    return p-s*clamp(round(p/s),lima,limb);
}


sdResult sceneSDF(vec3 sp) {
    vec3 q = sp-repeat*clamp(round(sp/repeat),-numRepeats,numRepeats);


    sdResult res;
    res = sdMin(  
        sdXYPlane(sp,-2,2),
       sdBox(q-boxPos, vec3(1,1,3), 0)
        
        );
    return res;
}



struct DistRes {
   float dist;
   int steps;
};

DistRes shortestDistanceToSurface(vec3 eye, vec3 marchingDirection, float start, float end) {
    DistRes result;
    //result.closestDist = 
    float depth = start;
    int i;
    for (i = 0; i < MAX_MARCHING_STEPS; i++) {
        sdResult res = sceneSDF(eye + depth * marchingDirection); 
        float dist = res.distance;
        if (dist < EPSILON) {
			result.dist = depth;
            result.steps = i;
			return result;
        }
        depth += dist;
        if (depth >= end) {
            result.dist = end;
            result.steps = i;
            return result;
        }
    }
    result.dist = end;        
    result.steps = i;
    return result;
    
}
struct ShadowRes {
   float closest;
   int steps;
};

ShadowRes softshadow( in vec3 ro, in vec3 rd, float mint, float maxt, float k)
{
    ShadowRes result;
    result.steps = 0; 
    result.closest = 1.0;
    for( float t=mint; t<maxt; )
    {        
        result.steps++;
        sdResult distInfo = sceneSDF(ro + rd*t);
        float h = distInfo.distance;
        if( h<0.001 ){
            result.closest = 0;
            return result;
        }
        result.closest = min( result.closest, k*h/t );
        t += h;

    }
    return result;
}





vec3 estimateNormal(vec3 p) {
    return normalize(vec3(
        sceneSDF(vec3(p.x + EPSILON, p.y, p.z)).distance - sceneSDF(vec3(p.x - EPSILON, p.y, p.z)).distance,
        sceneSDF(vec3(p.x, p.y + EPSILON, p.z)).distance - sceneSDF(vec3(p.x, p.y - EPSILON, p.z)).distance,
        sceneSDF(vec3(p.x, p.y, p.z  + EPSILON)).distance - sceneSDF(vec3(p.x, p.y, p.z - EPSILON)).distance
    ));
}





vec3 phong(vec3 p, vec3 eye ){
    //Blinn Phong shading        
    vec3 N = estimateNormal(p);
    vec3 L = normalize(lightPos - p);
    vec3 V = normalize(eye - p);
    vec3 R = normalize(reflect(-L, N));
    
    float dotLN = dot(L, N);
    float dotRV = dot(R, V);
    
    vec3 k_s = vec3(1); //Specular color
    float alpha = 10; //Shinyness
    vec3 lightCol;
    vec3 lightIntensity = vec3(.4);
    if (dotLN < 0.0) {
        // Light not visible from this point on the surface
        lightCol = vec3(0.0, 0.0, 0.0);
    } else if (dotRV < 0.0) {
        // Light reflection in opposite direction as viewer, apply only diffuse
        // component
        lightCol = lightIntensity  * dotLN;
    } else {
        lightCol = lightIntensity * dotLN + k_s * pow(dotRV, alpha);
    }

    return lightCol;
}



//uv should be x[-1,1] and y[-1,1]

vec3 rayDirection(float fieldOfView, vec2 uv) {
    vec2 xy = uv;
    float z = 1 / tan(radians(fieldOfView) / 2.0);
    return normalize(vec3(xy, -z));
}


mat4 viewMatrix(vec3 eye, vec3 center, vec3 up) {
	vec3 f = normalize(center - eye);
	vec3 s = normalize(cross(f, up));
	vec3 u = cross(s, f);
	return mat4(
		vec4(s, 0.0),
		vec4(u, 0.0),
		vec4(-f, 0.0),
		vec4(0.0, 0.0, 0.0, 1)
	);
}

void main() {
    float aspect = 8.0/6.0;
    vec2 uv = UV/vec2(aspect, 1);

    
	vec3 dir = rayDirection(45.0, uv);
    vec3 eye = camPos;
    
        
    mat4 viewToWorld = viewMatrix(eye, lookAt, vec3(0.0, 1.0, 0.0));
    vec3 worldDir = (viewToWorld * vec4(dir, 0.0)).xyz;
    
    
    DistRes Res = shortestDistanceToSurface(eye, worldDir, MIN_DIST, MAX_DIST);
    
    float dist = Res.dist;
    
    if (dist > MAX_DIST - EPSILON) {
        // Didn't hit anything
        if (DBDRAW!=0)
            frag_color = vec4(clearColor, 1.0);
		return;
    }


    vec3 colors[3];
    colors[0] = vec3(1,0,0);
    colors[1] = vec3(1,0,1);
    colors[2] = vec3(.5);
    // The closest point on the surface to the eyepoint along the view ray
    vec3 p = eye + dist * worldDir;

            
    int matID = sceneSDF(p).materialID;
   
   
    vec3 normal = estimateNormal(p);
   
   //Shadow start to avoid immediatly dieing
    vec3 sstart = p + normal*shadowEpsilon;
    ShadowRes shadow = softshadow(sstart, normalize(lightPos-sstart), 0 , length(sstart-lightPos), shadowK);

    
    vec3 color = colors[matID];

    vec3 lightCol = phong(p, eye);

    color += lightCol;
    color = color * clamp(shadow.closest+ambient, 0, 1);



    color = pow(color,vec3(gamma));


    frag_color = vec4(color, 1.0);


}




